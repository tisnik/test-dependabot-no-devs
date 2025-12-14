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

// CalcMandelbrotZ4 calculates Mandelbrot set z=z^4+c into the provided ZPixels
// Calculations use complex numbers
func CalcMandelbrotZ4(
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
				z = z*z*z*z + c
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += 3.0 / float64(image.Resolution.Width)
		}
		cy += 3.0 / float64(image.Resolution.Height)
	}
}
