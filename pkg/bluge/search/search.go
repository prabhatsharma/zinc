/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package search

import (
	"context"
	"fmt"
	"github.com/blugelabs/bluge/search/collector"
	"github.com/zinclabs/zinc/pkg/config"
	"sync/atomic"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
	"github.com/zinclabs/zinc/pkg/meta"
	"github.com/zinclabs/zinc/pkg/uquery"
	"golang.org/x/sync/errgroup"
)

func MultiSearch(ctx context.Context, query *meta.ZincQuery, mappings *meta.Mappings, analyzers map[string]*analysis.Analyzer, readers ...*bluge.Reader) (search.DocumentMatchIterator, error) {
	if len(readers) == 0 {
		return &DocumentList{
			bucket: search.NewBucket("",
				map[string]search.Aggregation{
					"duration": aggregations.Duration(),
				},
			),
		}, nil
	}
	if len(readers) == 1 {
		req, err := uquery.ParseQueryDSL(query, mappings, analyzers)
		if err != nil {
			return nil, err
		}
		return readers[0].Search(ctx, req)
	}

	bucketAggs := make(map[string]search.Aggregation)
	bucketAggs["duration"] = aggregations.Duration()

	eg := &errgroup.Group{}
	eg.SetLimit(config.Global.ReadGorutineNum)
	docs := make(chan *search.DocumentMatch, len(readers)*10)
	//aggsChan := make(chan *search.Bucket, len(readers))

	docList := &DocumentList{}
	egm := &errgroup.Group{}
	egm.Go(func() error {
		for doc := range docs {
			docList.bucket.Consume(doc)
			docList.addDocument(doc)
		}
		return nil
	})
	//egm.Go(func() error {
	//	for agg := range aggsChan {
	//		docList.bucket.Merge(agg)
	//	}
	//	return nil
	//})

	var sort search.SortOrder
	var size int
	var skip int
	var reversed bool
	var aggs search.Aggregations
	for _, r := range readers {
		req, err := uquery.ParseQueryDSL(query, mappings, analyzers)
		if err != nil {
			return nil, err
		}
		if sort == nil { // init vars
			aggs = req.Aggregations()
			docList.bucket = search.NewBucket("", aggs)
			sort = req.SortOrder()
			size, skip, reversed = req.SizeSkipAndReversed()
		}
		r := r
		eg.Go(func() error {
			var n int64
			dmi, err := r.Search(ctx, req)
			if err != nil {
				return err
			}
			next, err := dmi.Next()
			for err == nil && next != nil {
				n++
				docs <- next
				next, err = dmi.Next()
			}
			//aggsChan <- dmi.Aggregations()

			if n > atomic.LoadInt64(&docList.size) {
				atomic.StoreInt64(&docList.size, n)
			}

			fmt.Println("r", dmi.Aggregations().Duration().Milliseconds())

			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	close(docs)
	//close(aggsChan)
	_ = egm.Wait()

	err := docList.Done(size, skip, len(readers), reversed, sort)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

type DocumentList struct {
	docs   []*search.DocumentMatch
	bucket *search.Bucket
	size   int64
	next   int64
}

func (d *DocumentList) addDocument(doc *search.DocumentMatch) {
	d.docs = append(d.docs, doc)
}

func (d *DocumentList) Done(size, skip, numSearchers int, reversed bool, sort search.SortOrder) error {
	store := collector.NewCollectorStore(size, skip, reversed, sort)

	d.bucket.Finish()
	backingSize := size + skip + 1
	backingSize *= numSearchers

	var lowestMatchOutsideResults *search.DocumentMatch
	var removed *search.DocumentMatch
	for _, d := range d.docs {

		// optimization, we track lowest sorting hit already removed from heap
		// with this one comparison, we can avoid all heap operations if
		// this hit would have been added and then immediately removed
		if lowestMatchOutsideResults != nil {
			cmp := sort.Compare(d, lowestMatchOutsideResults)
			if cmp >= 0 {
				continue
			}
		}

		removed = store.AddNotExceedingSize(d, backingSize)
		if removed != nil {
			if lowestMatchOutsideResults == nil {
				lowestMatchOutsideResults = removed
			} else {
				cmp := sort.Compare(removed, lowestMatchOutsideResults)
				if cmp < 0 {
					lowestMatchOutsideResults = removed
				}
			}
		}
	}

	results, err := store.Final(skip, func(doc *search.DocumentMatch) error {
		doc.Complete(nil)
		return nil
	})
	if err != nil {
		return err
	}

	if reversed {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	d.docs = results
	return nil
}

func (d *DocumentList) Next() (*search.DocumentMatch, error) {
	if d.next >= d.size || d.next >= int64(len(d.docs)) {
		return nil, nil
	}
	doc := d.docs[d.next]
	d.next++
	return doc, nil
}

func (d *DocumentList) Aggregations() *search.Bucket {
	return d.bucket
}