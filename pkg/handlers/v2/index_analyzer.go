package v2

import (
	"fmt"
	"net/http"

	"github.com/blugelabs/bluge/analysis"
	"github.com/gin-gonic/gin"

	"github.com/prabhatsharma/zinc/pkg/core"
	zincanalysis "github.com/prabhatsharma/zinc/pkg/uquery/v2/analysis"
	"github.com/prabhatsharma/zinc/pkg/zutils"
)

func Analyze(c *gin.Context) {
	var query analyzeRequest
	if err := c.BindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	var ana *analysis.Analyzer
	indexName := c.Param("target")
	if indexName != "" {
		// use index analyzer
		index, exists := core.GetIndex(indexName)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "index " + indexName + " does not exists"})
			return
		}
		if query.Filed != "" && query.Analyzer == "" {
			if index.CachedMappings != nil && index.CachedMappings.Properties != nil {
				if prop, ok := index.CachedMappings.Properties[query.Filed]; ok {
					query.Analyzer = prop.Analyzer
				}
			}
		}
		ana, err = zincanalysis.QueryAnalyzer(index.CachedAnalysis, query.Analyzer)
		if err != nil {
			if query.Analyzer == "" {
				ana = new(analysis.Analyzer)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "analyzer " + query.Analyzer + " does not exists"})
				return
			}
		}
	} else {
		// none index specified
		ana, err = zincanalysis.QueryAnalyzer(nil, query.Analyzer)
		if err != nil {
			if query.Analyzer == "" {
				ana = new(analysis.Analyzer)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "analyzer " + query.Analyzer + " does not exists"})
				return
			}
		}
	}

	charFilters, err := parseCharFilter(query.CharFilter)
	if err != nil {
		handleError(c, err)
		return
	}

	tokenFilters, err := parseTokenFilter(query.TokenFilter)
	if err != nil {
		handleError(c, err)
		return
	}

	tokenizers, err := parseTokenizer(query.Tokenizer)
	if err != nil {
		handleError(c, err)
		return
	}

	if len(charFilters) > 0 {
		ana.CharFilters = append(ana.CharFilters, charFilters...)
	}

	if len(tokenFilters) > 0 {
		ana.TokenFilters = append(ana.TokenFilters, tokenFilters...)
	}

	if len(tokenizers) > 0 {
		ana.Tokenizer = tokenizers[0]
	}

	if ana.Tokenizer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "analyzer need set a tokenizer"})
		return
	}

	tokens := ana.Analyze([]byte(query.Text))
	ret := make([]gin.H, 0, len(tokens))
	for _, token := range tokens {
		ret = append(ret, formatToken(token))
	}
	c.JSON(http.StatusOK, gin.H{"tokens": ret})
}

func parseCharFilter(data interface{}) ([]analysis.CharFilter, error) {
	if data == nil {
		return nil, nil
	}

	chars := make([]analysis.CharFilter, 0)
	switch v := data.(type) {
	case string:
		filter, err := zincanalysis.RequestCharFilterSingle(v, nil)
		if err != nil {
			return nil, err
		}
		chars = append(chars, filter)
	case []interface{}:
		filters, err := zincanalysis.RequestCharFilterSlice(v)
		if err != nil {
			return nil, err
		}
		chars = append(chars, filters...)
	case map[string]interface{}:
		typ, err := zutils.GetStringFromMap(v, "type")
		if typ != "" && err == nil {
			filter, err := zincanalysis.RequestCharFilterSingle(typ, v)
			if err != nil {
				return nil, err
			}
			chars = append(chars, filter)
		} else {
			filters, err := zincanalysis.RequestCharFilter(v)
			if err != nil {
				return nil, err
			}
			for _, filter := range filters {
				chars = append(chars, filter)
			}
		}
	default:
		return nil, fmt.Errorf("char_filter unsuported type")
	}

	return chars, nil
}

func parseTokenFilter(data interface{}) ([]analysis.TokenFilter, error) {
	if data == nil {
		return nil, nil
	}

	tokens := make([]analysis.TokenFilter, 0)
	switch v := data.(type) {
	case string:
		filter, err := zincanalysis.RequestTokenFilterSingle(v, nil)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, filter)
	case []interface{}:
		filters, err := zincanalysis.RequestTokenFilterSlice(v)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, filters...)
	case map[string]interface{}:
		typ, err := zutils.GetStringFromMap(v, "type")
		if typ != "" && err == nil {
			filter, err := zincanalysis.RequestTokenFilterSingle(typ, v)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, filter)
		} else {
			filters, err := zincanalysis.RequestTokenFilter(v)
			if err != nil {
				return nil, err
			}
			for _, filter := range filters {
				tokens = append(tokens, filter)
			}
		}
	default:
		return nil, fmt.Errorf("token_filter unsuported type")
	}

	return tokens, nil
}

func parseTokenizer(data interface{}) ([]analysis.Tokenizer, error) {
	if data == nil {
		return nil, nil
	}

	tokenizers := make([]analysis.Tokenizer, 0)
	switch v := data.(type) {
	case string:
		zer, err := zincanalysis.RequestTokenizerSingle(v, nil)
		if err != nil {
			return nil, err
		}
		tokenizers = append(tokenizers, zer)
	case []interface{}:
		zers, err := zincanalysis.RequestTokenizerSlice(v)
		if err != nil {
			return nil, err
		}
		tokenizers = append(tokenizers, zers...)
	case map[string]interface{}:
		typ, err := zutils.GetStringFromMap(v, "type")
		if typ != "" && err == nil {
			zer, err := zincanalysis.RequestTokenizerSingle(typ, v)
			if err != nil {
				return nil, err
			}
			tokenizers = append(tokenizers, zer)
		} else {
			zers, err := zincanalysis.RequestTokenizer(v)
			if err != nil {
				return nil, err
			}
			for _, zer := range zers {
				tokenizers = append(tokenizers, zer)
			}
		}
	default:
		return nil, fmt.Errorf("tokenizer unsuported type")
	}

	return tokenizers, nil
}

func formatToken(token *analysis.Token) gin.H {
	return gin.H{
		"token":        string(token.Term),
		"start_offset": token.Start,
		"end_offset":   token.End,
		"position":     token.PositionIncr,
		"type":         formatTokenType(token.Type),
		"keyword":      token.KeyWord,
	}
}

func formatTokenType(typ analysis.TokenType) string {
	switch typ {
	case analysis.AlphaNumeric:
		return "<ALPHANUM>"
	case analysis.Ideographic:
		return "<IDEOGRAPHIC>"
	case analysis.Numeric:
		return "<NUM>"
	case analysis.DateTime:
		return "<DATETIME>"
	case analysis.Shingle:
		return "<SHINGLE>"
	case analysis.Single:
		return "<SINGLE>"
	case analysis.Double:
		return "<DOUBLE>"
	case analysis.Boolean:
		return "<BOOLEAN>"
	default:
		return "Unknown"
	}
}

type analyzeRequest struct {
	Analyzer    string      `json:"analyzer"`
	Filed       string      `json:"field"`
	Text        string      `json:"text"`
	Tokenizer   interface{} `json:"tokenizer"`
	CharFilter  interface{} `json:"char_filter"`
	TokenFilter interface{} `json:"token_filter"`
}
