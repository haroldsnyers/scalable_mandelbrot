package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math/cmplx"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Foo struct {
	Bar string
}

const (// à changer en fct du ask
	maxEsc = 30
	rMin   = -2.
	rMax   = .5
	iMin   = -1.
	iMax   = 1.
	width  = 7000
)

var foo []Foo

var palette []color.RGBA
var escapeColor color.RGBA
var wgx sync.WaitGroup

var scale float64
var height int
var bounds image.Rectangle

var b *image.NRGBA

func getMbrot(w http.ResponseWriter, req *http.Request) {
	scale = width / (rMax - rMin)
	height = int(scale * (iMax - iMin))
	bounds = image.Rect(0, 0, width, height)

	b = image.NewNRGBA(bounds)

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

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, b, nil); err != nil {
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

	log.Fatal(http.ListenAndServe(":8091", nil))
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
	wgx.Add(width)
	for x := 0; x < width; x++ {
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

	f, _ := os.Create("horizontal_scalability/slave/mandelbrot_.png")

	err := png.Encode(f, b)

	if err != nil {
		log.Println("png.Encode:", err)
	}

	if err = png.Encode(f, b); err != nil {
		fmt.Println(err)
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
}