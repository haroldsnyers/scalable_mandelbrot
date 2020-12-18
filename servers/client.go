package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type InfoServers struct {
	Name string `json:"name"`
	Port  string    `json:"port"`
}

var listServers []InfoServers

func main() {
	var resp *http.Response
	var err error

	resp, err = http.Get("http://localhost:8090/get_computation")

	if err != nil {
		fmt.Printf("Proxy on port %d not up", 8090)
		return
	}

	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&listServers)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			fmt.Printf("Bad Response. Wrong Type provided for field "+ unmarshalErr.Field, http.StatusBadRequest)
		} else {
			fmt.Printf("Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	fmt.Printf("Success", http.StatusOK)

	json_data, err := json.Marshal(listServers)

	if err != nil {

		log.Fatal(err)
	}

	fmt.Printf("\n%s", string(json_data))
}
