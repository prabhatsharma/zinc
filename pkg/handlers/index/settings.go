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

package index

import (
	"net/http"

	"github.com/blugelabs/bluge/analysis"
	"github.com/gin-gonic/gin"

	"github.com/zinclabs/zinc/pkg/core"
	"github.com/zinclabs/zinc/pkg/meta"
	zincanalysis "github.com/zinclabs/zinc/pkg/uquery/analysis"
)

// @Summary Get index settings
// @Tags    Index
// @Produce json
// @Param   target path  string  true  "Index"
// @Success 200 {object} map[string]meta.IndexSettings
// @Failure 400 {object} meta.HTTPResponse
// @Router /api/:target/_settings [get]
func GetSettings(c *gin.Context) {
	indexName := c.Param("target")
	index, exists := core.GetIndex(indexName)
	if !exists {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "index " + indexName + " does not exists"})
		return
	}

	settings := index.Settings
	if settings == nil {
		settings = new(meta.IndexSettings)
	}

	c.JSON(http.StatusOK, gin.H{index.Name: gin.H{"settings": settings}})
}

// @Summary Set index Settings
// @Tags    Index
// @Produce json
// @Param   target   path  string             true  "Index"
// @Param   settings body  meta.IndexSettings true  "Settings"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponse
// @Failure 500 {object} meta.HTTPResponse
// @Router /api/:target/_settings [put]
func SetSettings(c *gin.Context) {
	indexName := c.Param("target")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "index.name should be not empty"})
		return
	}

	var settings *meta.IndexSettings
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: err.Error()})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, meta.HTTPResponse{Message: "ok"})
		return
	}

	analyzers, err := zincanalysis.RequestAnalyzer(settings.Analysis)
	if err != nil {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: err.Error()})
		return
	}

	index, exists := core.GetIndex(indexName)
	if exists {
		// it can only change settings.NumberOfReplicas when index exists
		if settings.NumberOfReplicas > 0 {
			index.Settings.NumberOfReplicas = settings.NumberOfReplicas
		}
		if settings.Analysis != nil && len(settings.Analysis.Analyzer) > 0 {
			c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "can't update analyzer for existing index"})
			return
		}
		// store index
		if err := core.StoreIndex(index); err != nil {
			c.JSON(http.StatusInternalServerError, meta.HTTPResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, meta.HTTPResponse{Message: "ok"})
		return
	}

	var defaultSearchAnalyzer *analysis.Analyzer
	if analyzers != nil {
		defaultSearchAnalyzer = analyzers["default"]
	}
	index, err = core.NewIndex(indexName, "", defaultSearchAnalyzer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, meta.HTTPResponse{Error: err.Error()})
		return
	}

	// update settings
	_ = index.SetSettings(settings)

	// update analyzers
	_ = index.SetAnalyzers(analyzers)

	// store index
	if err := core.StoreIndex(index); err != nil {
		c.JSON(http.StatusInternalServerError, meta.HTTPResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, meta.HTTPResponse{Message: "ok"})
}
