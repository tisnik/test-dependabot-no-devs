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

// CalcBarnsleyJuliaJ3 calculates Barnsley J3 Mandelbrot-like set
func CalcBarnsleyJuliaJ3(
	params params.FractalParameter,
	image deepimage.Image) {

	cx := params.Cx0
	cy := params.Cy0
	var zy0 float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = zx0
			var zy float64 = zy0
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
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}
