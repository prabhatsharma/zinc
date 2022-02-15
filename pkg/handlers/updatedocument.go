package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/prabhatsharma/zinc/pkg/core"
)

func UpdateDocument(c *gin.Context) {
	indexName := c.Param("target")
	query_id := c.Param("id") // ID for the document to be updated provided in URL path

	var doc map[string]interface{}
	c.BindJSON(&doc)

	docID := ""
	mintedID := false

	// If id field is present then use it, else create a new UUID and use it
	if id, ok := doc["_id"]; ok {
		docID = id.(string)
	} else if query_id != "" {
		docID = query_id
	}
	if docID == "" {
		docID = uuid.New().String() // Generate a new ID if ID was not provided
		mintedID = true
	}

	var err error
	// If the index does not exist, then create it
	index, exists := core.GetIndex(indexName)
	if !exists {
		index, err = core.NewIndex(indexName, "disk") // Create a new index with disk storage as default
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		core.ZINC_INDEX_LIST[indexName] = index

		// use template
		template, _ := core.UseTemplate(indexName)
		if template != nil && template.Template.Mappings != nil && len(template.Template.Mappings.Properties) > 0 {
			index.SetMappings(template.Template.Mappings)
		}
	}

	// doc, _ = flatten.Flatten(doc, "", flatten.DotStyle)
	err = index.UpdateDocument(docID, &doc, mintedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"id": docID})
	}
}
