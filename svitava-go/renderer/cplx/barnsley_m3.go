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

// CalcBarnsleyMandelbrotM3 calculates the Barnsley M3 Mandelbrot-like fractal and fills the provided image buffers.
// The image pixel grid is mapped to the complex plane with real (x) and imaginary (y) coordinates spanning from -2 to +2.
// For each pixel the function iterates a piecewise complex recurrence (branch chosen by the sign of the current zx)
// until the iteration count reaches params.Maxiter or zx^2+zy^2 exceeds params.Bailout. The resulting complex value
// is written to image.Z and the iteration-based index (via calcIndex) is written to image.I.
// params supplies fractal parameters such as Maxiter and Bailout; image is the destination deepimage.Image to populate.
func CalcBarnsleyMandelbrotM3(
	params params.FractalParameter,
	image deepimage.Image) {

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = cx
			var zy float64 = cy
			var i uint
			for i < params.Maxiter {
				var zxn float64
				var zyn float64
				zx2 := zx * zx
				zy2 := zy * zy
				if zx2+zy2 > float64(params.Bailout) {
					break
				}
				if zx > 0 {
					zxn = zx2 - zy2 - 1
					zyn = 2.0 * zx * zy
				} else {
					zxn = zx2 - zy2 - 1 + cx*zx
					zyn = 2.0*zx*zy + cy*zx
				}
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}