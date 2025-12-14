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
	"math"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcNewton calculates Newton fractal
func CalcNewton(
	params params.FractalParameter,
	image deepimage.Image) {

	const Epsilon = 0.001

	var RootX1 = 1.0
	var RootY1 = 0.0

	var RootX2 = -0.5
	var RootY2 = math.Sqrt(3) / 2

	var RootX3 = -0.5
	var RootY3 = -math.Sqrt(3) / 2

	stepX, stepY := getSteps(params, image)

	var zy0 float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			zx := zx0
			zy := zy0
			i := uint(0)

			for i < params.Maxiter {
				zx2 := zx * zx
				zy2 := zy * zy
				zxn := 2.0/3.0*zx + (zx2-zy2)/(3.0*(zx2*zx2+zy2*zy2+2.0*zx2*zy2))
				zyn := 2.0/3.0*zy - 2.0*zx*zy/(3.0*(zx2*zx2+zy2*zy2+2.0*zx2*zy2))
				zx = zxn
				zy = zyn
				if math.Hypot(zx-RootX1, zy-RootY1) < Epsilon {
					break
				}
				if math.Hypot(zx-RootX2, zy-RootY2) < Epsilon {
					i += 128
					break
				}
				if math.Hypot(zx-RootX3, zy-RootY3) < Epsilon {
					i += 192
					break
				}
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(calcIndex(params, i))
			zx0 += stepX
		}
		zy0 += stepY
	}
}
