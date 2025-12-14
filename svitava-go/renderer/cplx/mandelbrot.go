//
//  (C) Copyright 2024, 2025  Pavel Tisnovsky
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

// CalcMandelbrot computes the Mandelbrot set over the region and resolution specified by params and image.
// It writes the final complex value for each pixel into image.Z and the iteration-based index into image.I.
func CalcMandelbrot(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = params.Cx0
			var zy float64 = params.Cy0
			var i uint
			for i < params.Maxiter {
				zx2 := zx * zx
				zy2 := zy * zy
				if zx2+zy2 > float64(params.Bailout) {
					break
				}
				zy = 2.0*zx*zy + cy
				zx = zx2 - zy2 + cx
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += stepX
		}
		cy += stepY
	}
}

// CalcMandelbrotComplex calculates Mandelbrot set into the provided ZPixels
// CalcMandelbrotComplex computes the Mandelbrot set for the region specified by params and writes the results into image.
// It evaluates z_{n+1} = z_n^2 + c using complex arithmetic per pixel, iterating until the iteration limit or the bailout condition is reached.
// For each pixel it stores the final complex value in image.Z and the iteration-derived index (via calcIndex) in image.I.
func CalcMandelbrotComplex(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			c := complex(cx, cy)
			z := complex(params.Cx0, params.Cy0)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > float64(params.Bailout) {
					break
				}
				z = z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += stepX
		}
		cy += stepY
	}
}

// CalcMandelbrotZ2pZ calculates Mandelbrot set z=z^2+z+c into the provided ZPixels
// CalcMandelbrotZ2pZ computes the fractal produced by the iteration z = z*z + z + c
// over a rectangular sampling of the complex plane and writes results into the image.
// The plane is sampled from -1.5 to +1.5 on both axes with step sizes 3/width and 3/height.
// Each pixel uses c = complex(cx, cy) and starts z at (params.Cx0, params.Cy0); iterations
// stop when |z|^2 > 4.0 or when params.Maxiter is reached. Final complex values are stored
// in image.Z and iteration-based indices (via calcIndex) are stored in image.I.
func CalcMandelbrotZ2pZ(
	params params.FractalParameter,
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
				z = z*z + z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}

// CalcMandelbrotZ2mZ calculates Mandelbrot set z=z^2-z+c into the provided ZPixels
// CalcMandelbrotZ2mZ computes the fractal defined by the iteration z = z*z - z + c
// and writes per-pixel results into the provided image.
//
// For each pixel the complex parameter c is sampled on a 3×3 region with real and
// imaginary components ranging from -1.5 to +1.5 mapped linearly to image resolution.
// The iterate z is initialized to (params.Cx0, params.Cy0) and advanced up to
// params.Maxiter iterations, stopping early when |z|^2 > 4.0. The final complex
// value is stored in image.Z and the iteration-based index (calcIndex(params, i))
// is stored in image.I.
func CalcMandelbrotZ2mZ(
	params params.FractalParameter,
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
				z = z*z - z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}