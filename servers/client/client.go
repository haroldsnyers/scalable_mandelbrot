package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
)

type InfoServers struct {
	Name string `json:"name"`
	Port  string    `json:"port"`
}

var listServers []InfoServers

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

// Keep it DRY so don't have to repeat opening file and decode
func OpenAndDecode(filepath string) (image.Image, string, error) {
	imgFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()
	img, format, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}
	return img, format, nil
}

// Decode image.Image's pixel data into []*Pixel
func DecodePixelsFromImage(img image.Image, offsetX, offsetY int) []*Pixel {
	pixels := []*Pixel{}
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			p := &Pixel{
				Point: image.Point{x + offsetX, y + offsetY},
				Color: img.At(x, y),
			}
			pixels = append(pixels, p)
		}
	}
	return pixels
}
//interactions with the slaves
func get(port string, name string, id int,total int,wg *sync.WaitGroup){
	fmt.Printf("server : %s", name)
	data := url.Values {
		"total": {strconv.Itoa(total)},
		"id": {strconv.Itoa(id)},
	}
	resp[id], err[id] = http.PostForm("http://localhost:"+ port +"/get_mbrot",data)
	defer wg.Done()

}

func check() bool{
	for i:=0; i< len(err);i ++{
		if err[i]!= nil{
			return false
		}
	}
	return true
}

func main() {
	getConnectedServers()
	generateMandelBrot()
}

func generateMandelBrot() {
	var wg sync.WaitGroup
	id := 0
	total := len(listServers)
	for i := 0; i< total;i++ {
		wg.Add(1)
		go get(listServers[i].Port, listServers[i].Name, id, total, &wg)
		id++
	}
	wg.Wait()


	if check(){
		for i:=0;i< total;i++{

			data[i], errRead[i] = ioutil.ReadAll(resp[i].Body)

			if errRead[i] != nil {
				log.Fatalf("ioutil.ReadAll -> %v", errRead[i])
			}
			_ = resp[i].Body.Close()

			picture[i], _, errImg[i] = image.Decode(bytes.NewReader(data[i]))
			if errImg[i] != nil {
				panic(errImg[i])
			}
		}

		out, _ := os.Create("horizontal_scalability/img.jpeg")
		defer out.Close()

		var opts jpeg.Options
		opts.Quality = 100

		errImg[0] = jpeg.Encode(out, picture[0], &opts)
		//jpeg.Encode(out, img, nil)
		if errImg[0] != nil {
			log.Println(errImg[0])
		}

		for i:=1; i< total;i++{
			picture[i], _, errImg[i] = image.Decode(bytes.NewReader(data[i]))
			if errImg[i] != nil {
				panic(errImg[i])
			}
		}

		// collect pixel data from each image

		pixels[0] = DecodePixelsFromImage(picture[0], 0, 0)
		pixelSum := append(pixels[0])
		lengthX := (picture[0].Bounds().Max.X)*total
		// the second image has a Y-offset of img1's max Y (appended at bottom)
		for i:=1;i<total;i++{
			pixels[i] = DecodePixelsFromImage(picture[i], (picture[i].Bounds().Max.X)*i, 0)
			pixelSum = append(pixelSum,pixels[i]...)
		}

		// Set a new size for the new image equal to the max width
		// of bigger image and max height of two images combined
		newRect := image.Rectangle{
			Min: picture[0].Bounds().Min,
			Max: image.Point{
				Y: picture[0].Bounds().Max.Y,
				X: lengthX,
			},
		}
		finImage := image.NewRGBA(newRect)
		// This is the cool part, all you have to do is loop through
		// each Pixel and set the image's color on the go
		for _, px := range pixelSum {
			finImage.Set(
				px.Point.X,
				px.Point.Y,
				px.Color,
			)
		}

		draw.Draw(finImage, finImage.Bounds(), finImage, image.Point{0, 0}, draw.Src)

		filename := "horizontal_scalability/output.png"
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
	} else {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
}

func getConnectedServers() {
	var resp *http.Response
	var err error

	resp, err = http.Get("http://localhost:8090/get_servers")

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


