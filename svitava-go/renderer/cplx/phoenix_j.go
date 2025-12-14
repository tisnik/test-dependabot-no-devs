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

// CalcPhoenixJ computes a Phoenix Julia–like fractal and writes per-pixel complex coordinates and intensity indices to the provided image.
// 
// It iterates every pixel in row-major order, performing an escape-time iteration using params (including Cx0, Cy0, Xmin, Ymin, Maxiter and Bailout).
// For each pixel it stores the final complex value in image.Z and the computed color/index (via calcIndex) in image.I. Iteration stops when the magnitude exceeds Bailout or when Maxiter is reached.
func CalcPhoenixJ(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	cx := params.Cx0
	cy := params.Cy0

	var zy0 float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = zx0
			var zy float64 = zy0
			var ynx = 0.0
			var yny = 0.0
			var i uint
			for i < params.Maxiter {
				zx2 := zx * zx
				zy2 := zy * zy
				zxn := zx2 - zy2 + cx + cy*ynx
				zyn := 2.0*zx*zy + cy*yny
				if zx2+zy2 > float64(params.Bailout) {
					break
				}
				ynx = zx
				yny = zy
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			zx0 += stepX
		}
		zy0 += stepY
	}
}