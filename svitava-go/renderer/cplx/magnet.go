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

// CalcMagnet computes a Magnet-style Mandelbrot set and writes the resulting complex values and iteration indices into the provided image.
// The function maps the complex plane region [-2,2] × [-2,2] across the image resolution, iterates each pixel up to params.Maxiter using
// params.Cx0 and params.Cy0 as the initial z, and stores the final complex value in image.Z and a corresponding index in image.I. The image
// is modified in place.
func CalcMagnet(
	params params.FractalParameter,
	image deepimage.Image) {
	const MIN_VALUE = 1.0 - 100

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = params.Cx0
			var zy float64 = params.Cy0
			var i uint
			for i < params.Maxiter {
				var zxn float64
				var zyn float64
				zx2 := zx * zx
				zy2 := zy * zy
				if zx2+zy2 > 100.0 {
					break
				}
				if ((zx-1.0)*(zx-1.0) + zy*zy) < 0.001 {
					break
				}
				tzx := zx2 - zy2 + cx - 1
				tzy := 2.0*zx*zy + cy
				bzx := 2.0*zx + cx - 2
				bzy := 2.0*zy + cy
				div := bzx*bzx + bzy*bzy
				if div < MIN_VALUE {
					break
				}
				zxn = (tzx*bzx + tzy*bzy) / div
				zyn = (tzy*bzx - tzx*bzy) / div
				zx = (zxn + zyn) * (zxn - zyn)
				zy = 2.0 * zxn * zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}