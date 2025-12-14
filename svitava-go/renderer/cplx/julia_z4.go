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

// CalcJuliaZ4 calculates the Julia fractal for the iteration z → z^4 + c over the provided image grid.
// It evaluates each pixel's orbit starting from the corresponding complex coordinate, stops when |z|^2 > 4 or when Maxiter is reached,
// and writes the final complex value to image.Z and the iteration-derived index to image.I.
// params supplies the coordinate bounds, the complex constant (Cx0,Cy0) and Maxiter; image is the destination buffer to populate.
func CalcJuliaZ4(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var zy0 float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(params.Cx0, params.Cy0)
			var z complex128 = complex(zx0, zy0)
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
			zx0 += stepX
		}
		zy0 += stepY
	}
}