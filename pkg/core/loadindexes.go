package core

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/prabhatsharma/zinc/pkg/zutils"
	"github.com/rs/zerolog/log"
)

var systemIndexList = []string{"_users", "_index_mapping"}

func LoadZincSystemIndexes() (map[string]*Index, error) {
	godotenv.Load()
	log.Print("Loading system indexes...")

	indexList := make(map[string]*Index)
	for _, systemIndex := range systemIndexList {

		tempIndex, err := NewIndex(systemIndex, "disk")
		if err != nil {
			log.Print("Error loading system index: ", systemIndex, " : ", err.Error())
			return nil, err
		}
		indexList[systemIndex] = tempIndex
		indexList[systemIndex].IndexType = "system"
		log.Print("Index loaded: " + systemIndex)
	}

	return indexList, nil
}

func LoadZincIndexesFromDisk() (map[string]*Index, error) {
	godotenv.Load()
	log.Print("Loading indexes... from disk")

	indexList := make(map[string]*Index)

	DATA_PATH := zutils.GetEnv("DATA_PATH", "./data")

	files, err := os.ReadDir(DATA_PATH)
	if err != nil {
		log.Print("Error reading data directory: ", err.Error())
		log.Fatal().Msg("Error reading data directory: " + err.Error())
	}

	for _, f := range files {
		iName := f.Name()

		iNameIsSystemIndex := false
		for _, systemIndex := range systemIndexList {
			if iName == systemIndex {
				iNameIsSystemIndex = true
			}
		}
		if iNameIsSystemIndex {
			continue
		}

		tempIndex, err := NewIndex(iName, "disk")
		if err != nil {
			log.Print("Error loading index: ", iName, " : ", err.Error()) // inform and move in to next index
		} else {
			indexList[iName] = tempIndex
			indexList[iName].IndexType = "user"
			log.Print("Index loaded: " + iName)
		}
	}

	return indexList, nil
}

func LoadZincIndexesFromS3() (map[string]*Index, error) {
	godotenv.Load()
	log.Print("Loading indexes from s3...")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Print("Error loading AWS config: ", err)
	}
	client := s3.NewFromConfig(cfg)

	IndexList := make(map[string]*Index)

	S3_BUCKET := zutils.GetEnv("S3_BUCKET", "")
	delimiter := "/"

	ctx := context.Background()
	params := s3.ListObjectsV2Input{
		Bucket:    &S3_BUCKET,
		Delimiter: &delimiter,
	}

	val, err := client.ListObjectsV2(ctx, &params)

	if err != nil {
		log.Print("failed to list indexes in s3: ", err.Error())
		return nil, err
	}

	for _, obj := range val.CommonPrefixes {

		iName := (*obj.Prefix)[0 : len(*obj.Prefix)-1]

		tempIndex, err := NewIndex(iName, "s3")

		if err != nil {
			log.Print("failed to load index "+iName+" in s3: ", err.Error())
		} else {
			IndexList[iName] = tempIndex
			IndexList[iName].IndexType = "user"
			IndexList[iName].StorageType = "s3"
			log.Print("Index loaded: " + iName)
		}

	}

	return IndexList, nil
}
