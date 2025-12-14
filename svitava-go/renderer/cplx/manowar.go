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

// CalcManowarM renders a Manowar Mandelbrot-like fractal into the provided image buffers.
// 
// For each pixel the function maps the pixel to a complex point c, iterates the recurrence
// z_{n+1} = z_n^2 + z_{n-1} + c starting from z_0 = z_1 = c, and stops when |z|^2 > 4 or when
// the iteration limit is reached. The final complex value is written into image.Z and the
// iteration-derived index from calcIndex is written into image.I.
func CalcManowarM(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(cx, cy)
			var z = c
			var z1 = c
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z2 := z*z + z1 + c
				z1 = z
				z = z2
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += stepX
		}
		cy += stepY
	}
}