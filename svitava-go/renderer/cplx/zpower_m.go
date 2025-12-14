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
	"math/cmplx"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcZPowerMandelbrot calculates Mandelbrot set into the provided ZPixels
// CalcZPowerMandelbrot computes a Mandelbrot-like fractal over the image using a z^z + z*z + c iteration.
// 
// For each pixel the function maps the pixel to the complex plane using params' bounds and the image resolution,
// initializes z and c to that complex coordinate, iterates z = z^z + z*z + c up to params.Maxiter or until the
// bailout condition is met, and writes the final complex value into image.Z and the iteration-based index
// (via calcIndex) into image.I.
func CalcZPowerMandelbrot(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			c := complex(cx, cy)
			z := complex(cx, cy)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > float64(params.Bailout) {
					break
				}
				z = cmplx.Pow(z, z) + z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += stepX
		}
		cy += stepY
	}
}