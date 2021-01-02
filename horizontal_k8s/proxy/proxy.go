package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

func getServers(w http.ResponseWriter, req *http.Request) {
	fmt.Print("Get Servers\n")
	fmt.Print(serverMap)

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

		url:= fmt.Sprintf("http://%s:%s/up", value, key)
		fmt.Print(url)
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
	http.HandleFunc("/get_servers", getServers)
	http.HandleFunc("/get_mbrot", getSubMandelbrot)

	fmt.Print("Serving ...\n")

	log.Fatal(http.ListenAndServe(":8090", nil))
}

func getSubMandelbrot(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	id, _ := strconv.Atoi(req.FormValue("id"))
	total, _ := strconv.Atoi(req.FormValue("total"))
	width := req.FormValue("width")
	maxEsc := req.FormValue("escape")
	port := req.FormValue("port")
	name := req.FormValue("server")

	data := url.Values {
		"total": {strconv.Itoa(total)},
		"id": {strconv.Itoa(id)},
		"width": {width},
		"escape": {maxEsc},
	}

	// nServer := strconv.Itoa(id + 1)

	resp, _ := http.PostForm("http://" + name + ":"+ port +"/get_mbrot",data)

	log.Printf("Read Bytes ...\n")
	dataRead, errRead := ioutil.ReadAll(resp.Body)

	if errRead != nil {
		log.Fatalf("ioutil.ReadAll -> %v", errRead)
	}

	picture, _, errImg := image.Decode(bytes.NewReader(dataRead))
	if errImg != nil {
		panic(errImg)
	}

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, picture, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}
