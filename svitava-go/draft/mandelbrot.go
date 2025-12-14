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

// CalcMandelbrotOneLine calculate one line of Mandelbrot set into the provided scanline of ZPixels
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
