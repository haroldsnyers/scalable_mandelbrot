package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// Create a struct to deal with pixel
type Pixel struct {
	Point image.Point
	Color color.Color
}

var resp1 *http.Response
var err1 error
var resp2 *http.Response
var err2 error

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
func get(port int, wg *sync.WaitGroup){
	if port == 8092{
		resp1, err1 = http.Get("http://localhost:8092/get_mbrot")
		defer wg.Done()
	}else{
		resp2, err2 = http.Get("http://localhost:8093/get_mbrot")
		defer wg.Done()
	}
}

func main() {
	//go routine for the slaves
	var wg sync.WaitGroup
	wg.Add(1)
	go get(8092, &wg)
	wg.Add(1)
	go get(8093, &wg)
	wg.Wait()

	if err1 == nil && err2 == nil{

		data1, errorRead1 := ioutil.ReadAll(resp1.Body)
		data2, errorRead2 := ioutil.ReadAll(resp2.Body)

		if errorRead1 != nil {
			log.Fatalf("ioutil.ReadAll -> %v", errorRead1)
		}

		if errorRead2 != nil {
			log.Fatalf("ioutil.ReadAll -> %v", errorRead2)
		}

		resp1.Body.Close()
		resp2.Body.Close()

		img1, _, err1 := image.Decode(bytes.NewReader(data1))
		if err1 != nil {
			panic(err1)
		}

		out, _ := os.Create("horizontal_scalability/img.jpeg")
		defer out.Close()

		var opts jpeg.Options
		opts.Quality = 100

		err1 = jpeg.Encode(out, img1, &opts)
		//jpeg.Encode(out, img, nil)
		if err1 != nil {
			log.Println(err1)
		}

		img2, _, err2 := image.Decode(bytes.NewReader(data2))
		if err2 != nil {
			panic(err2)
		}

		// collect pixel data from each image
		pixels1 := DecodePixelsFromImage(img1, 0, 0)
		// the second image has a Y-offset of img1's max Y (appended at bottom)
		pixels2 := DecodePixelsFromImage(img2, 0, img1.Bounds().Max.Y)
		pixelSum := append(pixels1, pixels2...)


		// Set a new size for the new image equal to the max width
		// of bigger image and max height of two images combined
		newRect := image.Rectangle{
			Min: img1.Bounds().Min,
			Max: image.Point{
				X: img2.Bounds().Max.X,
				Y: img2.Bounds().Max.Y + img1.Bounds().Max.Y,
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
		fmt.Printf("The HTTP request 1 failed with error %s\n", err1)
		fmt.Printf("The HTTP request 2 failed with error %s\n", err2)
	}
}