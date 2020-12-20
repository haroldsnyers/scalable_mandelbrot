package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/oliamb/cutter"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/cmplx"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var infoServer map[string]string

const (// à changer en fct du ask
	rMin   = -2.
	rMax   = .5
	iMin   = -1.2
	iMax   = 1.2
	// width = 7000
)

var width int
var maxEsc int

var palette []color.RGBA
var escapeColor color.RGBA
var wgx sync.WaitGroup

var scale float64
var height int
var bounds image.Rectangle

var b *image.NRGBA
var total int
var id int

func up(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	resp, _ := json.Marshal(infoServer)
	w.Write(resp)
}


func getMbrot(w http.ResponseWriter, req *http.Request) {
	log.Printf("Mandelbort Computation starting ... \n")

	query := req.URL.Query()
	widthGet := query.Get("width")
	maxEscGet := query.Get("escape")
	if widthGet == "" {
		width = 7000
	} else {
		width, _ = strconv.Atoi(widthGet)
	}
	if maxEscGet == "" {
		maxEsc = 30
	} else {
		maxEsc, _ = strconv.Atoi(maxEscGet)
	}

	_ = req.ParseForm()
	id, _ = strconv.Atoi(req.FormValue("id"))
	total, _ = strconv.Atoi(req.FormValue("total"))

	log.Printf("Mandelbrot generator : %d/%d\n", id + 1, total)

	scale = float64(width) / (rMax - rMin)
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
				if i % 5 == 0 {
					fmt.Print(".")
				}
				i++
			case <-done:
				ticker.Stop()
				fmt.Print("\n")
				log.Printf("Mandelbrot set rendered into `%s` in %d seconds", "mandelbrot_.png", i)
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


func main() {
	http.HandleFunc("/up", up)
	http.HandleFunc("/get_mbrot", getMbrot)

	var port int
	var name string

	port, _ = strconv.Atoi(os.Getenv("PORT"))
	name = os.Getenv("NAME")

	fmt.Printf("%s:%d", name, port)

	infoServer = map[string]string {
		"name": name,
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
	for x := id*(width/total); x <(id+1)*(width/total); x++ {
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
