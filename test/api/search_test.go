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

package api

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"

	"github.com/zinclabs/zinc/pkg/meta"
)

func TestSearch(t *testing.T) {

	t.Run("init data for search", func(t *testing.T) {
		body := bytes.NewBuffer(nil)
		body.WriteString(indexData)
		resp := request("PUT", "/api/"+indexName+"/_doc", body)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("POST /api/:target/_search", func(t *testing.T) {
		t.Run("search document with not exist indexName", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{}`)
			resp := request("POST", "/api/notExistSearch/_search", body)
			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
		t.Run("search document with exist indexName", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match_all":{}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)
		})
		t.Run("search document with not exist term", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match": {"_all": "xxxx"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.Equal(t, 0, data.Hits.Total.Value)
		})
		t.Run("search document with exist term", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match": {"_all": "DEMTSCHENKO"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: match_all", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match_all": {}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: wildcard", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"wildcard": {"_all": "dem*"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: fuzzy", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"fuzzy": {"Athlete": "demtschenk"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: term", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"term": {"City": "turin"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: daterange", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(
				fmt.Sprintf(`{"query": {"range": {"@timestamp": { "gte": "%s", "lt": "%s"}}}}`,
					time.Now().UTC().Add(time.Hour*-24).Format("2006-01-02T15:04:05Z"),
					time.Now().UTC().Format("2006-01-02T15:04:05Z"),
				))
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: match", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match": {"_all": "DEMTSCHENKO"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: matchphrase", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"match_phrase": {"_all": "DEMTSCHENKO"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: prefix", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"prefix": {"_all": "dem"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
		t.Run("search document type: querystring", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{"query": {"query_string": {"query": "DEMTSCHENKO"}}}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, data.Hits.Total.Value, 1)
		})
	})

	t.Run("POST /api/:target/_search with aggregations", func(t *testing.T) {
		t.Run("terms aggregation", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{
				"query": {"match_all":{}}, 
				"aggs": {
					"my-agg-term": {
						"terms": {"field": "City"}
					}
				}
			}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(data.Aggregations), 1)
		})

		t.Run("metric aggregation", func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.WriteString(`{
				"query": {"match_all":{}}, 
				"aggs": {
					"my-agg-max": {
						"max": {"field": "Year"}
					},
					"my-agg-min": {
						"min": {"field": "Year"}
					},
					"my-agg-avg": {
						"avg": {"field": "Year"}
					}
				}
			}`)
			resp := request("POST", "/api/"+indexName+"/_search", body)
			assert.Equal(t, http.StatusOK, resp.Code)

			data := new(meta.SearchResponse)
			err := json.Unmarshal(resp.Body.Bytes(), data)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(data.Aggregations), 1)
		})
	})
}
