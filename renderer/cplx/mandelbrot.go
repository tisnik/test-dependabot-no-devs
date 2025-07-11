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

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcMandelbrot computes the Mandelbrot set for each pixel using explicit real and imaginary parts and stores the results in the provided image.
// 
// For each pixel, maps its coordinates to a point in the complex plane, iterates z = z*z + c up to a maximum iteration count or until the escape condition is met, and records the final complex value and iteration count in the image.
func CalcMandelbrot(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -1.5
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = params.Cx0
			var zy float64 = params.Cy0
			var i uint
			for i < params.Maxiter {
				zx2 := zx * zx
				zy2 := zy * zy
				if zx2+zy2 > 4.0 {
					break
				}
				zy = 2.0*zx*zy + cy
				zx = zx2 - zy2 + cx
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}

// CalcMandelbrotComplex calculates Mandelbrot set into the provided ZPixels
// CalcMandelbrotComplex computes the Mandelbrot set for each pixel using Go's complex128 type and stores the results in the provided image.
// 
// For each pixel, the function maps its coordinates to a point in the complex plane, iterates z = z*z + c up to a maximum number of iterations or until the magnitude exceeds 2, and records the final complex value and iteration count in the image. The initial value of z is set from the provided parameters.
func CalcMandelbrotComplex(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -1.5
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(cx, cy)
			var z complex128 = complex(params.Cx0, params.Cy0)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z = z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}

// CalcMandelbrotZ3 calculates Mandelbrot set z=z^3+c into the provided ZPixels
// CalcMandelbrotZ3 computes a Mandelbrot set variant using the cubic iteration z = z³ + c for each pixel in the image.
// For each pixel, it maps coordinates to the complex plane, iterates up to a maximum count or until escape, and stores the final complex value and iteration count in the image.
func CalcMandelbrotZ3(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -1.5
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -1.5
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(cx, cy)
			var z complex128 = complex(params.Cx0, params.Cy0)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z = z*z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}

// CalcMandelbrotZ4 calculates Mandelbrot set z=z^4+c into the provided ZPixels
// CalcMandelbrotZ4 computes the Mandelbrot set using the quartic iteration z = z⁴ + c and stores the results in the provided image.
// For each pixel, it maps coordinates to the complex plane, iterates up to the maximum iteration count or until escape, and records the final complex value and iteration count.
func CalcMandelbrotZ4(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -1.5
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -1.5
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(cx, cy)
			var z complex128 = complex(params.Cx0, params.Cy0)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z = z*z*z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}
