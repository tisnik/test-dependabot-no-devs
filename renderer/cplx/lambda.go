//
//  (C) Copyright 2025  Pavel Tisnovsky
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

// CalcMandelLambda computes the Mandelbrot variant of the Lambda fractal over a specified region of the complex plane and stores the results in the provided image.
// 
// For each pixel in the image, the function maps its coordinates to a complex number, iteratively applies the Lambda fractal formula, and records both the final complex value and the iteration count at which the escape condition is met. The output is written to the image's Z and I pixel arrays.
func CalcMandelLambda(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -2.5
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
				z = c * z * (1 - z)
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += 6.0 / float64(image.Resolution.Width)
		}
		cy += 5.0 / float64(image.Resolution.Height)
	}
}
