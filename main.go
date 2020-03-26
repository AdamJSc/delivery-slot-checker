package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const asdaStatusUnavailable = "UNAVAILABLE"

type asdaResponse struct {
	StatusCode string `json:"statusCode"`
}

func main() {
	contents := getFailDataFromCache()
	var response asdaResponse

	err := json.Unmarshal(contents, &response)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode == asdaStatusUnavailable {
		log.Fatal("site is down")
	}

	log.Printf("%+v", response)
}

func getFailDataFromCache() []byte {
	data, err := ioutil.ReadFile("./data/cached_fail.txt")
	if err != nil {
		log.Fatal(err)
	}

	return data
}
