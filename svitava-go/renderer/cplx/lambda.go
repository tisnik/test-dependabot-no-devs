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

// CalcMandelLambda computes a lambda-family Mandelbrot fractal over the image grid and writes per-pixel complex values and iteration indices into the provided image.
// For each pixel, the complex constant c is formed from the pixel coordinates; iteration starts at (Cx0, Cy0) and applies z = c*z*(1-z) until reaching Maxiter or |z|^2 > 4.
// The final complex value is stored in image.Z and the iteration-based index (via calcIndex) is stored in image.I.
func CalcMandelLambda(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
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
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += stepX
		}
		cy += stepY
	}
}