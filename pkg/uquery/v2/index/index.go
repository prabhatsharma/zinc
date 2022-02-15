package index

import (
	"fmt"
	"strings"

	meta "github.com/prabhatsharma/zinc/pkg/meta/v2"
	"github.com/prabhatsharma/zinc/pkg/uquery/v2/mappings"
)

func Request(data map[string]interface{}) (*meta.Index, error) {
	if data == nil {
		return nil, nil
	}

	index := meta.NewIndex()
	for k, v := range data {
		k = strings.ToLower(k)
		switch k {
		case "settings":
			v, ok := v.(map[string]interface{})
			if !ok {
				return nil, meta.NewError(meta.ErrorTypeParsingException, "[index] settings should be an object")
			}
			for k, v := range v {
				k = strings.ToLower(k)
				switch k {
				case "number_of_replicas":
					index.Settings.NumberOfReplicas = int(v.(float64))
				case "number_of_shards":
					index.Settings.NumberOfShards = int(v.(float64))
				default:
					return nil, meta.NewError(meta.ErrorTypeParsingException, fmt.Sprintf("[index] settings unknown option [%s]", k))
				}
			}
		case "mappings":
			v, ok := v.(map[string]interface{})
			if !ok {
				return nil, meta.NewError(meta.ErrorTypeParsingException, "[index] mappings should be an object")
			}
			mappings, err := mappings.Request(v)
			if err != nil {
				return nil, err
			}
			index.Mappings = mappings
		default:
			return nil, meta.NewError(meta.ErrorTypeParsingException, fmt.Sprintf("[index] unknown option [%s]", k))
		}
	}

	return index, nil
}
