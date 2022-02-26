package analysis

import (
	"fmt"
	"strings"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/analyzer"

	"github.com/prabhatsharma/zinc/pkg/errors"
	meta "github.com/prabhatsharma/zinc/pkg/meta/v2"
	zincanalyzer "github.com/prabhatsharma/zinc/pkg/uquery/v2/analysis/analyzer"
)

func RequestAnalyzer(data *meta.IndexAnalysis) (map[string]*analysis.Analyzer, error) {
	if data == nil {
		return nil, nil
	}

	if data.Analyzer == nil {
		return nil, nil
	}

	charFilters, err := RequestCharFilter(data.CharFilter)
	if err != nil {
		return nil, err
	}

	if data.TokenFilter == nil && data.Filter != nil {
		data.TokenFilter = data.Filter
		data.Filter = nil
	}
	tokenFilters, err := RequestTokenFilter(data.TokenFilter)
	if err != nil {
		return nil, err
	}

	tokenizers, err := RequestTokenizer(data.Tokenizer)
	if err != nil {
		return nil, err
	}

	analyzers := make(map[string]*analysis.Analyzer)
	for name, v := range data.Analyzer {
		if v.Tokenizer == "" && v.Type == "" {
			return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] [%s] is missing tokenizer", name))
		}

		// custom build-in analyzer
		var ana *analysis.Analyzer
		if v.Type != "" {
			v.Type = strings.ToLower(v.Type)
			switch v.Type {
			case "custom":
				// omit
			case "regexp", "pattern":
				ana, err = zincanalyzer.NewRegexpAnalyzer(map[string]interface{}{
					"pattern":   v.Pattern,
					"lowercase": v.Lowercase,
					"stopwords": v.Stopwords,
				})
			case "standard":
				ana, err = zincanalyzer.NewStandardAnalyzer(map[string]interface{}{
					"stopwords": v.Stopwords,
				})
			case "stop":
				ana, err = zincanalyzer.NewStopAnalyzer(map[string]interface{}{
					"stopwords": v.Stopwords,
				})
			case "whitespace":
				ana, _ = zincanalyzer.NewWhitespaceAnalyzer()
			case "keyword":
				ana = analyzer.NewKeywordAnalyzer()
			case "simple":
				ana = analyzer.NewSimpleAnalyzer()
			case "web":
				ana = analyzer.NewWebAnalyzer()
			default:
				return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] build-in [%s] doesn't support custom", v.Type))
			}
			if err != nil {
				return nil, err
			}
		}

		// use tokenizer
		var ok bool
		zer, err := RequestTokenizerSingle(v.Tokenizer, nil)
		if zer != nil && err == nil {
			// use standard tokenizer
		} else {
			if zer, ok = tokenizers[v.Tokenizer]; !ok {
				if ana == nil { // returns error if not user build-in analyzer
					return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] [%s] used undifined tokenizer %s", name, v.Tokenizer))
				}
			}
		}

		chars := make([]analysis.CharFilter, 0, len(v.CharFilter))
		for _, filterName := range v.CharFilter {
			filter, err := RequestCharFilterSingle(filterName, nil)
			if filter != nil && err == nil {
				chars = append(chars, filter)
			} else {
				if v, ok := charFilters[filterName]; ok {
					chars = append(chars, v)
				} else {
					return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] [%s] used undefined char_filter [%s]", name, filterName))
				}
			}
		}

		tokens := make([]analysis.TokenFilter, 0, len(v.TokenFilter))
		if v.TokenFilter == nil && v.Filter != nil {
			v.TokenFilter = v.Filter
			v.Filter = nil
		}
		for _, filterName := range v.TokenFilter {
			filter, err := RequestTokenFilterSingle(filterName, nil)
			if filter != nil && err == nil {
				tokens = append(tokens, filter)
			} else {
				if v, ok := tokenFilters[filterName]; ok {
					tokens = append(tokens, v)
				} else {
					return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] [%s] used undefined token_filter [%s]", name, filterName))
				}
			}
		}

		if ana == nil {
			ana = &analysis.Analyzer{Tokenizer: zer}
		}
		if len(chars) > 0 {
			ana.CharFilters = append(ana.CharFilters, chars...)
		}
		if len(tokens) > 0 {
			ana.TokenFilters = append(ana.TokenFilters, tokens...)
		}
		analyzers[name] = ana
	}

	return analyzers, nil
}

func QueryAnalyzer(data map[string]*analysis.Analyzer, name string) (*analysis.Analyzer, error) {
	if name == "" {
		name = "default"
	}

	if data != nil {
		if v, ok := data[name]; ok {
			return v, nil
		}
	}

	switch name {
	case "standard":
		return zincanalyzer.NewStandardAnalyzer(nil)
	case "simple":
		return analyzer.NewSimpleAnalyzer(), nil
	case "keyword":
		return analyzer.NewKeywordAnalyzer(), nil
	case "web":
		return analyzer.NewWebAnalyzer(), nil
	case "regexp", "pattern":
		return zincanalyzer.NewRegexpAnalyzer(nil)
	case "stop":
		return zincanalyzer.NewStopAnalyzer(nil)
	case "whitespace":
		return zincanalyzer.NewWhitespaceAnalyzer()
	default:
		return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] [%s] doesn't exists", name))
	}
}

func QueryAnalyzerForField(data map[string]*analysis.Analyzer, mappings *meta.Mappings, field string) (*analysis.Analyzer, *analysis.Analyzer) {
	if field == "" {
		return nil, nil
	}

	analyzerName := ""
	searchAnalyzerName := ""
	if mappings != nil && len(mappings.Properties) > 0 {
		if v, ok := mappings.Properties[field]; ok {
			if v.Type != "text" {
				return nil, nil
			}
			if v.Analyzer != "" {
				analyzerName = v.Analyzer
			}
			if v.SearchAnalyzer != "" {
				searchAnalyzerName = v.SearchAnalyzer
			}
		}
	}

	analyzer, _ := QueryAnalyzer(data, analyzerName)
	searchAnalyzer, _ := QueryAnalyzer(data, searchAnalyzerName)

	return analyzer, searchAnalyzer
}
