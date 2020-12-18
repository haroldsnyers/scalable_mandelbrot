package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var port int

var infoServer map[string]string

func up(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	resp, _ := json.Marshal(infoServer)
	w.Write(resp)
}

func main() {
	http.HandleFunc("/up", up)
	port = 8091

	infoServer = map[string]string {
		"name": "server1",
		"port": fmt.Sprintf("%d", port),
	}

	reqBody, err := json.Marshal(infoServer)
	if err != nil {
		print(err)
	}
	resp, err := http.Post("http://proxy.resolute:8090/prox_connected",
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	fmt.Println(string(body))


	addr:= fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))

}
