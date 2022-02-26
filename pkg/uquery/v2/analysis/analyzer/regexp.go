package analyzer

import (
	"fmt"
	"regexp"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"

	zinctoken "github.com/prabhatsharma/zinc/pkg/bluge/analysis/token"
	"github.com/prabhatsharma/zinc/pkg/errors"
	"github.com/prabhatsharma/zinc/pkg/zutils"
)

func NewRegexpAnalyzer(options interface{}) (*analysis.Analyzer, error) {
	pattern, _ := zutils.GetStringFromMap(options, "pattern")
	if pattern == "" {
		pattern = "\\w+"
	}
	lowerCase, err := zutils.GetBoolFromMap(options, "lowercase")
	if err != nil {
		lowerCase = true
	}
	stopwords, _ := zutils.GetStringSliceFromMap(options, "stopwords")

	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[analyzer] regexp option [pattern] compile error: %v", err.Error()))
	}

	ana := &analysis.Analyzer{Tokenizer: tokenizer.NewRegexpTokenizer(r)}
	if lowerCase {
		ana.TokenFilters = append(ana.TokenFilters, token.NewLowerCaseFilter())
	}

	if len(stopwords) > 0 {
		ana.TokenFilters = append(ana.TokenFilters, zinctoken.NewStopTokenFilter(stopwords))
	}

	return ana, nil
}
