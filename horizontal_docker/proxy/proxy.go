package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type InfoServers struct {
	Name string `json:"name"`
	Port  string    `json:"port"`
}

var serverMap map[string]string
var list []InfoServers

func up(w http.ResponseWriter, req *http.Request) {
	// Double check it's a post request being made
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid_http_method")
		return
	}
	var e InfoServers
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&e)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			errorResponse(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	json_data, err := json.Marshal(e)

	if err != nil {

		log.Fatal(err)
	}

	errorResponse(w, "Successfully registered to proxy", http.StatusOK)
	fmt.Print("server " + string(json_data) + " connected \n")

	if serverMap == nil {
		serverMap = make(map[string]string)
	}

	serverMap[e.Port] = e.Name

	return
}


func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func getComputation(w http.ResponseWriter, req *http.Request) {
	fmt.Print("Get Computation\n")

	getStatus(w)

	json_data, err := json.Marshal(list)

	if err != nil {

		log.Fatal(err)
	}
	w.Write(json_data)
}

func getStatus(w http.ResponseWriter) {
	list = nil

	for key, value := range serverMap {

		var resp *http.Response
		var err error

		url:= fmt.Sprintf("http://%s.resolute:%s/up", value, key)
		resp, err = http.Get(url)

		if err != nil {
			delete(serverMap, key)
			fmt.Printf("server on port %s not up, deleting server ...", key)
		} else {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading response. ", err)
			}

			var e InfoServers
			var unmarshalErr *json.UnmarshalTypeError

			decoder := json.NewDecoder(bytes.NewReader(body))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&e)

			if err != nil {
				if errors.As(err, &unmarshalErr) {
					errorResponse(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
				} else {
					errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
				}
				return
			}

			list = append(list, e)
		}
	}

}

func main() {
	http.HandleFunc("/prox_connected", up)
	http.HandleFunc("/get_servers", getComputation)

	fmt.Print("Serving ...\n")

	log.Fatal(http.ListenAndServe(":8090", nil))
}
