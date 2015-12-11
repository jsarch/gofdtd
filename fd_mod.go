package main

import (
	"fmt"
	// "runtime"
)

type Section struct {
	in, out      [][]float64
	i_dx2, i_dy2 float64
	row_0, row_n chan float64
	col_0, col_m chan float64
}

func (s *Section) swapInOut() {
	s.in, s.out = s.out, s.in
}

func (s *Section) Print() {
	print(s.in)
	print(s.out)
}

func print(in [][]float64) {
	for x := range in {
		for y := range in[x] {
			fmt.Printf("%f\t", in[x][y])
		}
		fmt.Println()
	}
	fmt.Println()
}

func convolve(section Section) {

	in := section.in
	out := section.out

	i_dx2, i_dy2 := section.i_dx2, section.i_dy2

	row := 0
	for col := 1; col < len(in[row])-1; col++ {
		in[row][col] = <-section.row_0
	}
	row = len(in) - 1
	for col := 1; col < len(in[row])-1; col++ {
		in[row][col] = <-section.row_n
	}

	for row := 1; row < len(in)-1; row++ {

		for col := 1; col < len(in[row])-1; col++ {

			out[row][col] = (in[row][col-1]+in[row][col+1]-2*in[row][col])*i_dx2 + (in[row-1][col]+in[row+1][col]-2*in[row][col])*i_dy2 + in[row][col]

		}

	}
}

func boundary_top(ch chan float64, dx int) {
	for {
		for col := 1; col < dx-1; col++ {
			ch <- 0.0
		}
	}
}

func boundary_bottom(ch chan float64, dx int) {
	for {
		for col := 1; col < dx-1; col++ {
			ch <- 0.0
		}
	}
}

func boundary_left(ch chan float64, dy int) {
	for {
		for row := 1; row < dy-1; row++ {
			ch <- 0.0
		}
	}
}

func boundary_right(ch chan float64, dy int) {
	for {
		for row := 1; row < dy-1; row++ {
			ch <- 0.0
		}
	}
}

func create(dx, dy int) [][]float64 {
	out := make([][]float64, dy)

	// Allocate one large slice to hold all the pixels.
	_out := make([]float64, dx*dy)
	for i := range out {
		out[i], _out = _out[:dx], _out[dx:]
	}

	return out
}

func run() {
	const (
		dx    = 4
		dy    = 5
		i_dx2 = 1.0 / float64(dx*dx)
		i_dy2 = 1.0 / float64(dy*dy)
	)

	section := Section{
		create(dx, dy),
		create(dx, dy),
		i_dx2, i_dy2,
		make(chan float64, dx), make(chan float64, dx),
		make(chan float64, dy), make(chan float64, dy)}

	section.in[1][2] = 3

	section.Print()

	go boundary_top(section.row_0, dx)
	go boundary_bottom(section.row_n, dx)
	go boundary_left(section.col_0, dy)
	go boundary_right(section.col_m, dy)

	convolve(section)

	section.Print()

	section.swapInOut()

	convolve(section)

	section.Print()

}

type Domain struct {
	in, out                                    [][]float64
	numRows, numCols                           int
	startRow, startCol                         int
	topLeft, topRight, bottomLeft, bottomRight *Domain
	totalRows, totalCols                       int
	toRight, toLeft, toTop, toBottom           []chan float64
	fromRight, fromLeft, fromTop, fromBottom   []chan float64
}

func (d *Domain) print() {
	if d.topLeft != nil && d.topRight != nil && d.bottomLeft != nil && d.bottomRight != nil {
		d.topLeft.print()
		d.topRight.print()
		d.bottomLeft.print()
		d.bottomRight.print()
	} else {
		for row := range d.in {
			for col := range d.in[row] {
				fmt.Printf("%f\t", d.in[row][col])
			}
			fmt.Println()
		}
		fmt.Println()

		for row := range d.out {
			for col := range d.out[row] {
				fmt.Printf("%f\t", d.out[row][col])
			}
			fmt.Println()
		}
		fmt.Println()

	}
}

func (d *Domain) init() {
	if d.topLeft != nil && d.topRight != nil && d.bottomLeft != nil && d.bottomRight != nil {
		d.topLeft.init()
		d.topRight.init()
		d.bottomLeft.init()
		d.bottomRight.init()
	} else {
		d.in = create(d.numCols+2, d.numRows+2)
		d.out = create(d.numCols+2, d.numRows+2)

		for i := 0; i < d.numRows; i++ {
			for j := 0; j < d.numCols; j++ {
				d.in[i+1][j+1] = float64((d.startRow+i)*d.totalCols + d.startCol + j)
			}
		}
	}
}

func (d *Domain) aliasDataRowsCols(parent *Domain) {
	d.in = make([][]float64, d.numRows)

	for i := range d.in {
		d.in[i] = parent.in[d.startRow+i][d.startCol : d.startCol+d.numCols]
	}
}

func (d *Domain) split2x2() {

	halfRows := d.numRows / 2
	halfCols := d.numCols / 2

	sendTLtoTR := make([]chan float64, halfRows)
	sendTRtoTL := make([]chan float64, halfRows)
	sendBLtoBR := make([]chan float64, halfRows)
	sendBRtoBL := make([]chan float64, halfRows)

	sendTLtoBL := make([]chan float64, halfCols)
	sendBLtoTL := make([]chan float64, halfCols)
	sendTRtoBR := make([]chan float64, halfCols)
	sendBRtoTR := make([]chan float64, halfCols)

	for i := 0; i < halfRows; i++ {
		sendTLtoTR[i] = make(chan float64, 1) // TL -> TR
		sendTRtoTL[i] = make(chan float64, 1) // TL <- TR
		sendBLtoBR[i] = make(chan float64, 1) // BL -> BR
		sendBRtoBL[i] = make(chan float64, 1) // BL <- BR
	}

	for i := 0; i < halfCols; i++ {
		sendTLtoBL[i] = make(chan float64, 1) // TL -> BL
		sendBLtoTL[i] = make(chan float64, 1) // TL <- BL
		sendTRtoBR[i] = make(chan float64, 1) // TR -> BR
		sendBRtoTR[i] = make(chan float64, 1) // TR <- BR
	}

	d.topLeft = &Domain{numRows: halfRows, numCols: halfCols, startRow: d.startRow, startCol: d.startCol, totalRows: d.totalRows, totalCols: d.totalCols,
		toRight: sendTLtoTR, fromRight: sendTRtoTL, toBottom: sendTLtoBL, fromBottom: sendBLtoTL, toLeft: nil, fromLeft: d.fromLeft[:halfRows], toTop: nil, fromTop: d.fromTop[:halfCols]}
	// d.topLeft.aliasDataRowsCols(d

	d.topRight = &Domain{numRows: halfRows, numCols: halfCols, startRow: d.startRow, startCol: d.startCol + halfCols, totalRows: d.totalRows, totalCols: d.totalCols,
		fromLeft: sendTLtoTR, toLeft: sendTRtoTL, toBottom: sendTRtoBR, fromBottom: sendBRtoTR, toRight: nil, fromRight: d.fromRight[:halfRows], toTop: nil, fromTop: d.fromTop[halfCols:]}
	// d.topRight.aliasDataRowsCols(d)

	d.bottomLeft = &Domain{numRows: halfRows, numCols: halfCols, startRow: d.startRow + halfRows, startCol: d.startCol, totalRows: d.totalRows, totalCols: d.totalCols,
		toRight: sendBLtoBR, fromRight: sendBRtoBL, fromTop: sendTLtoBL, toTop: sendBLtoTL, toLeft: nil, fromLeft: d.fromLeft[halfRows:], toBottom: nil, fromBottom: d.fromBottom[:halfCols]}
	// d.bottomLeft.aliasDataRowsCols(d)

	d.bottomRight = &Domain{numRows: halfRows, numCols: halfCols, startRow: d.startRow + halfRows, startCol: d.startCol + halfCols, totalRows: d.totalRows, totalCols: d.totalCols,
		fromLeft: sendBLtoBR, toLeft: sendBRtoBL, fromTop: sendTRtoBR, toTop: sendBRtoTR, toRight: nil, fromRight: d.fromRight[halfRows:], toBottom: nil, fromBottom: d.fromBottom[halfCols:]}
	// d.bottomRight.aliasDataRowsCols(d)

	if d.toLeft != nil {
		d.topLeft.toLeft = d.toLeft[:halfRows]
		d.bottomLeft.toLeft = d.toLeft[halfRows:]
	}
	if d.toRight != nil {
		d.topRight.toRight = d.toRight[:halfRows]
		d.bottomRight.toRight = d.toRight[halfRows:]
	}
	if d.toTop != nil {
		d.topLeft.toTop = d.toTop[:halfCols]
		d.topRight.toTop = d.toTop[halfCols:]
	}
	if d.toBottom != nil {
		d.bottomLeft.toBottom = d.toBottom[:halfCols]
		d.bottomRight.toBottom = d.toBottom[halfCols:]
	}

	if d.topLeft.numRows > 2 && d.topLeft.numCols > 2 {
		d.topLeft.split2x2()
	}
	if d.topRight.numRows > 2 && d.topRight.numCols > 2 {
		d.topRight.split2x2()
	}
	if d.bottomLeft.numRows > 2 && d.bottomLeft.numCols > 2 {
		d.bottomLeft.split2x2()
	}
	if d.bottomRight.numRows > 2 && d.bottomRight.numCols > 2 {
		d.bottomRight.split2x2()
	}

	// d.topLeft.print()
	// d.topRight.print()
	// d.bottomLeft.print()
	// d.bottomRight.print()

}

func leftBoundaryCondition(ch []chan float64, numRows int) {
	for {
		for row := 0; row < numRows; row++ {
			ch[row] <- -1.0
		}
	}
}

func rightBoundaryCondition(ch []chan float64, numRows int) {
	for {
		for row := 0; row < numRows; row++ {
			ch[row] <- -3.0
		}
	}
}

func topBoundaryCondition(ch []chan float64, numCols int) {
	for {
		for col := 0; col < numCols; col++ {
			ch[col] <- -2.0
		}
	}
}

func bottomBoundaryCondition(ch []chan float64, numCols int) {
	for {
		for col := 0; col < numCols; col++ {
			ch[col] <- -4.0
		}
	}
}

func (d *Domain) solve(c chan bool, iterations int) {
	if d.topLeft != nil && d.topRight != nil && d.bottomLeft != nil && d.bottomRight != nil {
		ch := make(chan bool, 4)

		go d.topLeft.solve(ch, iterations)
		go d.topRight.solve(ch, iterations)
		go d.bottomLeft.solve(ch, iterations)
		go d.bottomRight.solve(ch, iterations)

		for i := 0; i < 4; i++ {
			<-ch
		}
	} else {

		i_dx2 := 1.0 / float64(d.totalCols*d.totalCols)
		i_dy2 := 1.0 / float64(d.totalRows*d.totalRows)

		for i := 0; i < iterations; i++ {
			in := d.in
			out := d.out

			// Send to Right
			if d.toRight != nil {
				for row := 1; row < d.numRows+1; row++ {
					d.toRight[row-1] <- in[row][d.numCols]
				}
			}
			// Send to Left
			if d.toLeft != nil {
				for row := 1; row < d.numRows+1; row++ {
					d.toLeft[row-1] <- in[row][1]
				}
			}
			// Send to Top
			if d.toTop != nil {
				for col := 1; col < d.numCols+1; col++ {
					d.toTop[col-1] <- in[1][col]
				}
			}
			// Send to Bottom
			if d.toBottom != nil {
				for col := 1; col < d.numCols+1; col++ {
					d.toBottom[col-1] <- in[d.numRows][col]
				}
			}

			// Receive from Left
			if d.fromLeft != nil {
				for row := 1; row < d.numRows+1; row++ {
					in[row][0] = <-d.fromLeft[row-1]
				}
			}
			// Receive from Right
			if d.fromRight != nil {
				for row := 1; row < d.numRows+1; row++ {
					in[row][d.numCols+1] = <-d.fromRight[row-1]
				}
			}
			// Receive from Top
			if d.fromTop != nil {
				for col := 1; col < d.numCols+1; col++ {
					in[0][col] = <-d.fromTop[col-1]
				}
			}
			// Receive from Bottom
			if d.fromBottom != nil {
				for col := 1; col < d.numCols+1; col++ {
					in[d.numRows+1][col] = <-d.fromBottom[col-1]
				}
			}

			for row := 1; row < d.numRows+1; row++ {
				for col := 1; col < d.numCols+1; col++ {

					dx := (in[row][col-1] + in[row][col+1] - 2*in[row][col]) * i_dx2
					dy := (in[row-1][col] + in[row+1][col] - 2*in[row][col]) * i_dy2

					out[row][col] = dx + dy + in[row][col]

					// out[row][col] = (in[row][col] + in[row-1][col] + in[row+1][col] + in[row][col-1] + in[row][col+1]) / 5
				}
			}

			d.in, d.out = d.out, d.in
		}
	}

	c <- true
}

func main() {
	// runtime.GOMAXPROCS(2)

	// run()

	const (
		totalRows = 16
		totalCols = 8
	)

	// data := create(totalCols, totalRows)
	// for i := range data {
	// 	for j := range data[i] {
	// 		data[i][j] = float64(i*len(data[i]) + j)
	// 	}
	// }

	leftBoundary := make([]chan float64, totalRows)
	rightBoundary := make([]chan float64, totalRows)
	for row := 0; row < totalRows; row++ {
		leftBoundary[row] = make(chan float64, 1)
		rightBoundary[row] = make(chan float64, 1)
	}
	topBoundary := make([]chan float64, totalCols)
	bottomBoundary := make([]chan float64, totalCols)
	for col := 0; col < totalCols; col++ {
		topBoundary[col] = make(chan float64, 1)
		bottomBoundary[col] = make(chan float64, 1)
	}

	root := Domain{numRows: totalRows, numCols: totalCols, totalRows: totalRows, totalCols: totalCols,
		fromLeft: leftBoundary, fromRight: rightBoundary, fromTop: topBoundary, fromBottom: bottomBoundary}

	root.split2x2()

	root.init()

	// root.print()

	c := make(chan bool)

	go leftBoundaryCondition(leftBoundary, totalRows)
	go rightBoundaryCondition(rightBoundary, totalRows)
	go topBoundaryCondition(topBoundary, totalCols)
	go bottomBoundaryCondition(bottomBoundary, totalCols)

	go root.solve(c, 1)

	<-c

	root.print()

}
