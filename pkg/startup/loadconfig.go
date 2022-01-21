package startup

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var MAX_RESULTS = LoadMaxResults()

func LoadMaxResults() int {
	godotenv.Load()
	MAX_RESULTS_STRING := os.Getenv("MAX_RESULTS")

	if MAX_RESULTS_STRING != "" {
		MAX_RESULTS, err := strconv.Atoi(MAX_RESULTS_STRING)
		if err != nil {
			return 10000
		} else {
			return MAX_RESULTS
		}
	}

	return 10000
}
