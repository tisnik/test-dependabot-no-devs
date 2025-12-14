//
//  (C) Copyright 2024  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package cplx

type ZImage [][]ZPixel

// ZPixel is a representation of pixel in complex plane with additional metadata
type ZPixel struct {
	Iter uint
	Z    complex128
}

// CalcMandelbrotOneLine computes the Mandelbrot iteration results for a single horizontal scanline
// and writes a ZPixel for each column into zimageLine.
//
// pcx and pcy specify the initial real and imaginary parts of z for each pixel; cy is the imaginary
// part of the complex constant c used for the entire scanline while the real part of c is stepped
// across the line. maxiter bounds the iteration count. zimageLine is modified in place and must have
// length at least width. The height parameter is unused. Completion is signaled by sending true on done.
func CalcMandelbrotOneLine(width uint, height uint, pcx float64, pcy float64, maxiter uint, zimageLine []ZPixel, cy float64, done chan bool) {
	var cx float64 = -2.0
	for x := uint(0); x < width; x++ {
		var zx float64 = pcx
		var zy float64 = pcy
		var i uint
		for i < maxiter {
			zx2 := zx * zx
			zy2 := zy * zy
			if zx2+zy2 > 4.0 {
				break
			}
			zy = 2.0*zx*zy + cy
			zx = zx2 - zy2 + cx
			i++
		}
		zimageLine[x] = ZPixel{Iter: i, Z: complex(zx, zy)}
		cx += 3.0 / float64(width)
	}
	done <- true
}

// calcMandelbrotComplex computes Mandelbrot iterations for a single horizontal line at imaginary coordinate cy,
// maps iteration counts to RGB using palette, writes the resulting 3*width bytes into image, and signals completion on done.
// The image slice must have length at least 3*width and palette must provide a color for each possible iteration value up to maxiter.
func calcMandelbrotComplex(width uint, height uint, maxiter uint, palette [][3]byte, image []byte, cy float64, done chan bool) {
	var c complex128 = complex(-2.0, cy)
	var dc complex128 = complex(3.0/float64(width), 0.0)
	for x := uint(0); x < width; x++ {
		var z complex128 = 0.0 + 0.0i
		var i uint
		for i < maxiter {
			zx := real(z)
			zy := imag(z)
			if zx*zx+zy*zy > 4.0 {
				break
			}
			z = z*z + c
			i++
		}
		color := palette[i]
		image[3*x] = color[0]
		image[3*x+1] = color[1]
		image[3*x+2] = color[2]
		c += dc
	}
	done <- true
}