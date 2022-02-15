package core

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/jeremywohl/flatten"
	"github.com/rs/zerolog/log"

	meta "github.com/prabhatsharma/zinc/pkg/meta/v2"
)

// BuildBlugeDocumentFromJSON returns the bluge document for the json document. It also updates the mapping for the fields if not found.
// If no mappings are found, it creates te mapping for all the encountered fields. If mapping for some fields is found but not for others
// then it creates the mapping for the missing fields.
func (index *Index) BuildBlugeDocumentFromJSON(docID string, doc *map[string]interface{}) (*bluge.Document, error) {
	// Pick the index mapping from the cache if it already exists
	mappings := index.CachedMappings
	if mappings == nil {
		mappings = new(meta.Mappings)
		mappings.Properties = make(map[string]meta.Property)
	}

	mappingsNeedsUpdate := false

	// Create a new bluge document
	bdoc := bluge.NewDocument(docID)
	flatDoc, _ := flatten.Flatten(*doc, "", flatten.DotStyle)
	// Iterate through each field and add it to the bluge document
	for key, value := range flatDoc {
		if value == nil {
			continue
		}

		if _, ok := mappings.Properties[key]; !ok {
			// Use reflection to find the type of the value.
			// Bluge requires the field type to be specified.
			v := reflect.ValueOf(value)

			// try to find the type of the value and use it to define default mapping
			switch v.Type().String() {
			case "string":
				mappings.Properties[key] = meta.NewProperty("text")
			case "float64":
				mappings.Properties[key] = meta.NewProperty("numeric")
			case "bool":
				mappings.Properties[key] = meta.NewProperty("bool")
			case "time.Time":
				mappings.Properties[key] = meta.NewProperty("time")
			}

			mappingsNeedsUpdate = true
		}

		if !mappings.Properties[key].Index {
			continue // not index, skip
		}

		var field *bluge.TermField
		switch mappings.Properties[key].Type {
		case "text":
			field = bluge.NewTextField(key, value.(string)).SearchTermPositions()
		case "numeric":
			field = bluge.NewNumericField(key, value.(float64))
		case "keyword":
			// compatible verion <= v0.1.4
			if v, ok := value.(bool); ok {
				field = bluge.NewKeywordField(key, strconv.FormatBool(v))
			} else if v, ok := value.(string); ok {
				field = bluge.NewKeywordField(key, v).Aggregatable()
			} else {
				return nil, fmt.Errorf("keyword type only support text")
			}
		case "bool": // found using existing index mapping
			value := value.(bool)
			field = bluge.NewKeywordField(key, strconv.FormatBool(value))
		case "time":
			tim, err := time.Parse(time.RFC3339, value.(string))
			if err != nil {
				return nil, err
			}
			field = bluge.NewDateTimeField(key, tim)
		}

		if mappings.Properties[key].Store {
			field.StoreValue()
		}
		if mappings.Properties[key].Sortable {
			field.Sortable()
		}
		if mappings.Properties[key].Aggregatable {
			field.Aggregatable()
		}
		if mappings.Properties[key].Highlightable {
			field.HighlightMatches()
		}
		bdoc.AddField(field)
	}

	if mappingsNeedsUpdate {
		index.SetMappings(mappings)
	}

	docByteVal, _ := json.Marshal(*doc)
	bdoc.AddField(bluge.NewDateTimeField("@timestamp", time.Now()).StoreValue().Sortable().Aggregatable())
	bdoc.AddField(bluge.NewStoredOnlyField("_source", docByteVal))
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil)) // Add _all field that can be used for search

	return bdoc, nil
}

// SetMapping Saves the mapping of the index to _index_mapping index
// index: Name of the index ffor which the mapping needs to be saved
// iMap: a map of the fileds that specify name and type of the field. e.g. movietitle: string
func (index *Index) SetMappings(mappings *meta.Mappings) error {
	// @timestamp need date_range/date_histogram aggregation, and mappings used for type check in aggregation
	mappings.Properties["@timestamp"] = meta.NewProperty("time")

	bdoc := bluge.NewDocument(index.Name)
	for k, prop := range mappings.Properties {
		bdoc.AddField(bluge.NewTextField(k, prop.Type).StoreValue())
	}

	docByteVal, _ := json.Marshal(mappings)
	bdoc.AddField(bluge.NewStoredOnlyField("_source", docByteVal))
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))

	// update on the disk
	systemIndex := ZINC_SYSTEM_INDEX_LIST["_index_mapping"].Writer
	err := systemIndex.Update(bdoc.ID(), bdoc)
	if err != nil {
		log.Printf("error updating document: %v", err)
		return err
	}

	// update in the cache
	index.CachedMappings = mappings

	return nil
}

// GetStoredMapping returns the mappings of all the indexes from _index_mapping system index
func (index *Index) GetStoredMapping() (*meta.Mappings, error) {
	for _, indexName := range systemIndexList {
		if index.Name == indexName {
			return nil, nil
		}
	}

	reader, _ := ZINC_SYSTEM_INDEX_LIST["_index_mapping"].Writer.Reader()
	defer reader.Close()

	// search for the index mapping _index_mapping index
	query := bluge.NewTermQuery(index.Name).SetField("_id")
	searchRequest := bluge.NewTopNSearch(1, query) // Should get just 1 result at max
	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Error().Str("index", index.Name).Msg("error executing search: " + err.Error())
	}

	next, err := dmi.Next()
	if err != nil {
		return nil, err
	}

	mappings := new(meta.Mappings)
	oldMappings := make(map[string]string)
	if next != nil {
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case "_source":
				if string(value) != "" {
					json.Unmarshal(value, mappings)
				}
			default:
				oldMappings[field] = string(value)
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}

	// compatible old mappings format
	if len(oldMappings) > 0 && len(mappings.Properties) == 0 {
		mappings.Properties = make(map[string]meta.Property, len(oldMappings))
		for k, v := range oldMappings {
			mappings.Properties[k] = meta.NewProperty(v)
		}
	}

	if len(mappings.Properties) == 0 {
		mappings.Properties = make(map[string]meta.Property)
	}

	return mappings, nil
}
