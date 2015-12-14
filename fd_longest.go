package main

import (
	"fmt"
	// "runtime"
)

func print(in [][]float64) {
	for x := range in {
		for y := range in[x] {
			fmt.Printf("%f\t", in[x][y])
		}
		fmt.Println()
	}
	fmt.Println()
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

type Domain struct {
	in, out                                    [][]float64
	numRows, numCols                           int
	startRow, startCol                         int
	top, bottom, left, right                   *Domain
	totalRows, totalCols                       int
	toRight, toLeft, toTop, toBottom           []chan float64
	fromRight, fromLeft, fromTop, fromBottom   []chan float64
}

func (d *Domain) print() {
	if d.top != nil && d.bottom != nil {
		d.top.print()
		d.bottom.print()
	} else if d.left != nil && d.right != nil {
		d.left.print()
		d.right.print()
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
	if d.top != nil && d.bottom != nil {
		d.top.init()
		d.bottom.init()
	} else if d.left != nil && d.right != nil {
		d.left.init()
		d.right.init()
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

func (d *Domain) split() {
	halfRows := d.numRows / 2
	halfCols := d.numCols / 2

	if halfRows > halfCols && halfRows > 1 {
		d.splitRows()
		d.top.split()
		d.bottom.split()
	} else if halfCols > 1 {
		d.splitCols()
		d.left.split()
		d.right.split()
	}
}

func (d *Domain) splitRows() {
	sendToptoBottom := make([]chan float64, d.numCols)
	sendBottomtoTop := make([]chan float64, d.numCols)

	for i := 0; i < d.numCols; i++ {
		sendToptoBottom[i] = make(chan float64, 1) // Top -> Bottom
		sendBottomtoTop[i] = make(chan float64, 1) // Bottom -> Top
	}

	halfRows := d.numRows / 2

	d.top = &Domain{
		numRows:    halfRows,
		numCols:    d.numCols,
		startRow:   d.startRow,
		startCol:   d.startCol,
		totalRows:  d.totalRows,
		totalCols:  d.totalCols,
		toTop:      nil,
		fromTop:    d.fromTop,
		toBottom:   sendToptoBottom,
		fromBottom: sendBottomtoTop,
		toLeft:     nil,
		fromLeft:   d.fromLeft[:halfRows],
		toRight:    nil,
		fromRight:  d.fromRight[:halfRows],
	}

	d.bottom = &Domain{
		numRows:    halfRows,
		numCols:    d.numCols,
		startRow:   d.startRow + halfRows,
		startCol:   d.startCol,
		totalRows:  d.totalRows,
		totalCols:  d.totalCols,
		toTop:      sendBottomtoTop,
		fromTop:    sendToptoBottom,
		toBottom:   nil,
		fromBottom: d.fromBottom,
		toLeft:     nil,
		fromLeft:   d.fromLeft[halfRows:],
		toRight:    nil,
		fromRight:  d.fromRight[halfRows:],
	}

	if d.toLeft != nil {
		d.top.toLeft = d.toLeft[:halfRows]
		d.bottom.toLeft = d.toLeft[halfRows:]
	}
	if d.toRight != nil {
		d.top.toRight = d.toRight[:halfRows]
		d.bottom.toRight = d.toRight[halfRows:]
	}
	if d.toTop != nil {
		d.top.toTop = d.toTop
	}
	if d.toBottom != nil {
		d.bottom.toBottom = d.toBottom
	}

	// if d.top.numRows > 2 {
	// 	d.top.splitRows()
	// }

}

func (d *Domain) splitCols() {
	sendLefttoRight := make([]chan float64, d.numRows)
	sendRighttoLeft := make([]chan float64, d.numRows)

	for i := 0; i < d.numRows; i++ {
		sendLefttoRight[i] = make(chan float64, 1) // Left -> Right
		sendRighttoLeft[i] = make(chan float64, 1) // Right -> Left
	}

	halfCols := d.numCols / 2

	d.left = &Domain{
		numRows:    d.numRows,
		numCols:    halfCols,
		startRow:   d.startRow,
		startCol:   d.startCol,
		totalRows:  d.totalRows,
		totalCols:  d.totalCols,
		toTop:      nil,
		fromTop:    d.fromTop[:halfCols],
		toBottom:   nil,
		fromBottom: d.fromBottom[:halfCols],
		toLeft:     nil,
		fromLeft:   d.fromLeft,
		toRight:    sendLefttoRight,
		fromRight:  sendRighttoLeft,
	}

	d.right = &Domain{
		numRows:    d.numRows,
		numCols:    halfCols,
		startRow:   d.startRow,
		startCol:   d.startCol + halfCols,
		totalRows:  d.totalRows,
		totalCols:  d.totalCols,
		toTop:      nil,
		fromTop:    d.fromTop[halfCols:],
		toBottom:   nil,
		fromBottom: d.fromBottom[halfCols:],
		toLeft:     sendRighttoLeft,
		fromLeft:   sendLefttoRight,
		toRight:    nil,
		fromRight:  d.fromRight,
	}

	if d.toLeft != nil {
		d.left.toLeft = d.toLeft
	}
	if d.toRight != nil {
		d.right.toRight = d.toRight
	}
	if d.toTop != nil {
		d.left.toTop = d.toTop[:halfCols]
		d.right.toTop = d.toTop[halfCols:]
	}
	if d.toBottom != nil {
		d.left.toBottom = d.toBottom[:halfCols]
		d.right.toBottom = d.toBottom[halfCols:]
	}

	// if d.left.numCols > 2 {
	// 	d.left.splitCols()
	// }

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
	if d.top != nil && d.bottom != nil {
		ch := make(chan bool, 2)

		go d.top.solve(ch, iterations)
		go d.bottom.solve(ch, iterations)

		for i := 0; i < 2; i++ {
			<-ch
		}
	} else if d.left != nil && d.right != nil {
		ch := make(chan bool, 2)

		go d.left.solve(ch, iterations)
		go d.right.solve(ch, iterations)

		for i := 0; i < 2; i++ {
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
		totalRows = 128
		totalCols = 128
	)

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

	root.split()

	root.init()

	// root.print()

	c := make(chan bool)

	go leftBoundaryCondition(leftBoundary, totalRows)
	go rightBoundaryCondition(rightBoundary, totalRows)
	go topBoundaryCondition(topBoundary, totalCols)
	go bottomBoundaryCondition(bottomBoundary, totalCols)

	go root.solve(c, 100)

	<-c

	root.print()

}
