package main

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"time"
)

func print(b [][]float64) {
	for i := range b {
		for j := range b[i] {
			fmt.Printf("%f\t", b[i][j])
		}
		fmt.Println()
	}
	fmt.Println()
}

func initialize(b [][]float64, n int) {
	for i := 0; i < n; i++ {
		rand.Seed(time.Now().UTC().UnixNano())
		x, y, v := rand.Intn(cap(b)), rand.Intn(cap(b)), rand.Float64()*255
		fmt.Println(x, y, v)

		b[x][y] = v
	}
	// print(b)
}

func convolve(a, b [][]float64) {
	for i := 0; i < len(b); i++ {
		for j := 0; j < len(b[i]); j++ {
			if i == 0 || i == len(b[i])-1 || j == 0 || j == len(b[i])-1 {
				b[i][j] = a[i][j]
			} else {
				b[i][j] = (a[i][j] + a[i-1][j] + a[i+1][j] + a[i][j-1] + a[i][j+1]) / 5
			}
		}
	}

}

func main() {
	run()
}

func create1(dx, dy int) (a, b [][]float64) {
	a, b = make([][]float64, dy), make([][]float64, dy)

	for i := range a {
		a[i], b[i] = make([]float64, dx), make([]float64, dx)
	}

	return
}

func create2(dx, dy int) (a, b [][]float64) {
	a, b = make([][]float64, dy), make([][]float64, dy)

	// Allocate one large slice to hold all the pixels.
	_a, _b := make([]float64, dx*dy), make([]float64, dx*dy)
	for i := range a {
		a[i], _a = _a[:dx], _a[dx:]
	}
	for i := range b {
		b[i], _b = _b[:dx], _b[dx:]
	}

	return
}

func run() {
	const (
		dx = 1024
		dy = 1024
	)

	a, b := create2(dx, dy)

	initialize(a, 5)

	for i := 0; i < 100; i++ {
		// print(a)
		convolve(a, b)
		// print(b)
		convolve(b, a)
	}
	Show(a, dx, dy)
}

func Show(b [][]float64, dx, dy int) {
	max := 0.0
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			if b[y][x] > max {
				max = b[y][x]
			}
		}
	}

	m := image.NewNRGBA(image.Rect(0, 0, dx, dy))
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			v := uint8(b[y][x] / max * 255)
			// fmt.Println(v)
			i := y*m.Stride + x*4
			m.Pix[i] = v
			m.Pix[i+1] = v
			m.Pix[i+2] = 255
			m.Pix[i+3] = 255
		}
	}
	ShowImage(m)
}

func ShowImage(m image.Image) {

	w, _ := os.Create("tmp.png")
	defer w.Close()
	png.Encode(w, m) //Encode writes the Image m to w in PNG format.
}
