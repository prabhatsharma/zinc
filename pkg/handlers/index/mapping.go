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

	"github.com/gin-gonic/gin"

	"github.com/zinclabs/zinc/pkg/core"
	"github.com/zinclabs/zinc/pkg/meta"
	"github.com/zinclabs/zinc/pkg/uquery/mappings"
)

// @Summary Get index mappings
// @Tags    Index
// @Produce json
// @Param   target path  string  true  "Index"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} meta.HTTPResponse
// @Router /api/:target/_mapping [get]
func GetMapping(c *gin.Context) {
	indexName := c.Param("target")
	index, exists := core.GetIndex(indexName)
	if !exists {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "index " + indexName + " does not exists"})
		return
	}

	// format mappings
	mappings := index.Mappings
	if mappings == nil {
		mappings = meta.NewMappings()
	}

	c.JSON(http.StatusOK, gin.H{index.Name: gin.H{"mappings": mappings}})
}

// @Summary Set index mappings
// @Tags    Index
// @Produce json
// @Param   target  path  string        true  "Index"
// @Param   mapping body  meta.Mappings true  "Mapping"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponse
// @Failure 500 {object} meta.HTTPResponse
// @Router /api/:target/_mapping [put]
func SetMapping(c *gin.Context) {
	indexName := c.Param("target")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "index.name should be not empty"})
		return
	}

	var mappingRequest map[string]interface{}
	if err := c.BindJSON(&mappingRequest); err != nil {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: err.Error()})
		return
	}

	mappings, err := mappings.Request(nil, mappingRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: err.Error()})
		return
	}

	index, exists := core.GetIndex(indexName)
	if exists {
		// check if mapping field is exists
		if index.Mappings != nil && index.Mappings.Len() > 0 {
			for field := range mappings.ListProperty() {
				if _, ok := index.Mappings.GetProperty(field); ok {
					c.JSON(http.StatusBadRequest, meta.HTTPResponse{Error: "index [" + indexName + "] already exists mapping of field [" + field + "]"})
					return
				}
			}
		}
		// add mappings
		for field, prop := range mappings.ListProperty() {
			index.Mappings.SetProperty(field, prop)
		}
		mappings = index.Mappings
	} else {
		// create index
		index, err = core.NewIndex(indexName, "", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, meta.HTTPResponse{Error: err.Error()})
			return
		}
	}

	// update mappings
	if mappings != nil && mappings.Len() > 0 {
		_ = index.SetMappings(mappings)
	}

	// store index
	if err := core.StoreIndex(index); err != nil {
		c.JSON(http.StatusInternalServerError, meta.HTTPResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, meta.HTTPResponse{Message: "ok"})
}
