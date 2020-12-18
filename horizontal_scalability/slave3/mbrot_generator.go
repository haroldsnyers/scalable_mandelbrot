package main

import (
	"../src/github.com/oliamb/cutter"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math/cmplx"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (// à changer en fct du ask
	maxEsc = 30
	rMin   = -2.
	rMax   = .5
	iMin   = -1.
	iMax   = 1.
	width  = 7000
)

var palette []color.RGBA
var escapeColor color.RGBA
var wgx sync.WaitGroup

var scale float64
var height int
var bounds image.Rectangle

var b *image.NRGBA
var total int
var id int

func getMbrot(w http.ResponseWriter, req *http.Request) {
	scale = width / (rMax - rMin)
	height = int(scale * (iMax - iMin))
	bounds = image.Rect(0, 0, width, height)

	b = image.NewNRGBA(bounds)

	_ = req.ParseForm()
	id, _ = strconv.Atoi(req.FormValue("id"))
	total, _ = strconv.Atoi(req.FormValue("total"))

	done := make(chan struct{})
	ticker := time.NewTicker(time.Millisecond * 1000)
	go func() {
		i := 0
		for {
			select {
			case <-ticker.C:
				fmt.Print(".")
				i++
			case <-done:
				ticker.Stop()
				fmt.Printf("\n\nMandelbrot set rendered into `%s` in %d seconds\n", "mandelbrot_.png", i)
			}
		}
	}()

	render(done)

	croppedImg, _ := cutter.Crop(b, cutter.Config{
		Width:  width/total,
		Height: height,
		Anchor: image.Point{id*(width/total), 0},
	})

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, croppedImg, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

func mandelbrot(a complex128) int {
	z := a
	i := 0
	// must end absolute value of z is superior to 2 or we attain n max
	for ; i < maxEsc - 1 ; i++ {
		if cmplx.Abs(z) > 2 {
			return i
		}
		z = cmplx.Pow(z, 2) + a
	}
	return maxEsc-i
}

func main() {

	http.HandleFunc("/get_mbrot", getMbrot)

	log.Fatal(http.ListenAndServe(":8094", nil))
}

func render(done chan struct{}, ) {
	palette = make([]color.RGBA, maxEsc)
	for i := 0; i < maxEsc-1; i++ {
		palette[i] = color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			255}
	}
	escapeColor = color.RGBA{0, 0, 0, 0}

	//draw.Draw(b, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	wgx.Add(width/total)
	for x := id*(width/total); x < (id+1)*(width/total); x++ {
		go func(xx int) {
			defer wgx.Done()
			for y := 0; y < height; y++ {
				coord := complex(float64(xx)/scale+rMin, float64(y)/scale+iMin)
				fEsc := mandelbrot(coord)
				if fEsc == maxEsc - 1 {
					b.Set(xx, y, escapeColor)
				}
				b.Set(xx, y, palette[fEsc])

			}
		}(x)
	}
	wgx.Wait()
	done <- struct{}{}
}