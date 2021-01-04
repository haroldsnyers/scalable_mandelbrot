package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var proxyServer string

type InfoServers struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

var listServersUp []InfoServers

var finImage *image.RGBA

// Create a struct to deal with pixel
type Pixel struct {
	Point image.Point
	Color color.Color
}

var resp [100]*http.Response
var err [100]error
var errRead [100]error
var errImg [100]error
var data [100][]byte
var picture [100]image.Image
var pixels [100][]*Pixel

func main() {
	log.Printf("Starting ... ")
	width := 400
	escape := 30

	proxyServer = "127.0.0.1:33669"
	// proxyServer = "localhost:8090"

	getMandelbrot(strconv.Itoa(width), strconv.Itoa(escape))
}

func getMandelbrot(width string, escape string) {
	getConnectedServers()
	generateMandelBrot(width, escape)
}

func getConnectedServers() {
	var resp *http.Response
	var err error

	resp, err = http.Get("http://" + proxyServer + "/get_servers")

	if err != nil {
		log.Printf("Proxy on port %d not up", 8090)
		return
	}

	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&listServersUp)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			log.Printf("Bad Response. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			log.Printf("Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	log.Printf("Success, all horizontal_docker retrieved (status: %d)", http.StatusOK)
}

func generateMandelBrot(width string, escape string) {
	log.Printf("Mandelbort Computation starting ... \n")

	var wg sync.WaitGroup
	id := 0
	total := len(listServersUp)
	for i := 0; i < total; i++ {
		wg.Add(1)
		go getMandelbrotSub(listServersUp[i].Port, listServersUp[i].Name, id, total, &wg, width, escape)
		id++
	}
	wg.Wait()

	log.Printf("Mandelbrot data generated\n")
	log.Printf("Starting image generation ...\n")

	if checkForErrorsDuringPartGeneration() {
		for i := 0; i < total; i++ {
			data[i], errRead[i] = ioutil.ReadAll(resp[i].Body)

			if errRead[i] != nil {
				log.Fatalf("ioutil.ReadAll -> %v", errRead[i])
			}
			_ = resp[i].Body.Close()
			log.Printf("image decoding part %d...\n", i + 1)
			picture[i], _, errImg[i] = image.Decode(bytes.NewReader(data[i]))
			if errImg[i] != nil {
				panic(errImg[i])
			}
		}

		log.Printf("Images data decoded ... \n")

		out, _ := os.Create("horizontal_k8s/images/img.jpeg")
		defer out.Close()

		//Using a part of the picture to know the boundary
		var opts jpeg.Options
		opts.Quality = 100
		errImg[0] = jpeg.Encode(out, picture[0], &opts)
		if errImg[0] != nil {
			log.Println(errImg[0])
		}

		// collect pixel data from each image
		log.Printf("Collecting pixel data from each image ...\n")
		pixels[0] = DecodePixelsFromImage(picture[0], 0, 0)
		pixelSum := append(pixels[0])
		lengthX := (picture[0].Bounds().Max.X) * total
		// the second image has a Y-offset of img1's max Y (appended at bottom)
		for i := 1; i < total; i++ {
			pixels[i] = DecodePixelsFromImage(picture[i], (picture[i].Bounds().Max.X)*i, 0)
			pixelSum = append(pixelSum, pixels[i]...)
		}

		log.Printf("Creating full image to fit sum of pixels ... \n")
		// Set a new size for the new image equal to the max width
		// of bigger image and max height of two images combined
		newRect := image.Rectangle{
			Min: picture[0].Bounds().Min,
			Max: image.Point{
				Y: picture[0].Bounds().Max.Y,
				X: lengthX,
			},
		}

		log.Printf("Adding colors to image ...\n")
		finImage = image.NewRGBA(newRect)
		// This is the cool part, all you have to do is loop through
		// each Pixel and set the image's color on the go
		for _, px := range pixelSum {
			finImage.Set(
				px.Point.X,
				px.Point.Y,
				px.Color,
			)
		}
		log.Printf("Drawing Image ... \n")
		draw.Draw(finImage, finImage.Bounds(), finImage, image.Point{}, draw.Src)

		saveImage()
	} else {
		log.Printf("The HTTP request failed with error %s\n", err)
	}
}

// Decode image.Image's pixel data into []*Pixel
func DecodePixelsFromImage(img image.Image, offsetX, offsetY int) []*Pixel {
	var pixels []*Pixel
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			p := &Pixel{
				Point: image.Point{X: x + offsetX, Y: y + offsetY},
				Color: img.At(x, y),
			}
			pixels = append(pixels, p)
		}
	}
	return pixels
}

//interactions with the slaves via proxy
func getMandelbrotSub(port string, name string, id int, total int, wg *sync.WaitGroup, width string, escape string) {
	log.Printf("Send computation request to %s\n", name)

	if width == "" {
		width = "4000"
	}
	if escape == "" {
		escape = "30"
	}

	data := url.Values{
		"total":  {strconv.Itoa(total)},
		"id":     {strconv.Itoa(id)},
		"width":  {width},
		"escape": {escape},
		"port":   {port},
		"server": {name},
	}

	resp[id], err[id] = http.PostForm("http://"+proxyServer+"/get_mbrot", data)

	defer wg.Done()
}

func checkForErrorsDuringPartGeneration() bool {
	for i := 0; i < len(err); i++ {
		if err[i] != nil {
			return false
		}
	}
	return true
}

func saveImage() {
	log.Printf("Saving image ...\n")
	dt := time.Now()
	//Format MM-DD-YYYY hh:mm:ss
	date := dt.Format("01-02-2006T15-04-05")

	filename := "horizontal_k8s/images/output" + date + ".png"
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
