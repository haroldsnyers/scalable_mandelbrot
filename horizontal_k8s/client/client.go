package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var finImage image.Image

func main() {
	width := 4000
	escape := 30

	getImage(width, escape)
}

func getImage(width int, escape int) {
	var resp *http.Response
	var err error

	minikubePort := "127.0.0.1:39547" // external port of proxy in minikube cluster
	// minikubePort = "localhost:8089"

	data := url.Values {
		"width": {strconv.Itoa(width)},
		"escape": {strconv.Itoa(escape)},
	}

	resp, err = http.PostForm("http://" + minikubePort +"/get_mbrot", data)

	dataRead, errRead := ioutil.ReadAll(resp.Body)

	if errRead != nil {
		log.Fatalf("ioutil.ReadAll -> %v", errRead)
	}
	_ = resp.Body.Close()

	finImage, _,  err = image.Decode(bytes.NewReader(dataRead))
	if err != nil {
		fmt.Println("Error decoding", err) // "unknown format"
	}
	saveImage()

}

func saveImage() {
	log.Printf("Saving image ...\n")
	dt := time.Now()
	//Format MM-DD-YYYY hh:mm:ss
	date := dt.Format("01-02-2006T15-04-05")

	filename := "horizontal_docker/images/output" + date + ".png"
	// Create a new file and write to it
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, finImage)
	if err != nil {
		panic(err)
		os.Exit(1)
	}

	log.Printf("I saved your image (%s) buddy!\n", filename)
}
