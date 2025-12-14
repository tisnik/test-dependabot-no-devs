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

package attractors_2d

import (
	"math"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// jason_rampe_2 computes the next point of the Jason–Rampe II attractor for the given coordinates and parameters.
// It returns (xn, yn) where xn = cos(b*y) + c*cos(b*x) and yn = cos(a*x) + d*cos(a*y).
func jason_rampe_2(x, y, a, b, c, d float64) (float64, float64) {
	xn := math.Cos(b*y) + c*math.Cos(b*x)
	yn := math.Cos(a*x) + d*math.Cos(a*y)
	return xn, yn
}

// CalcJasonRampe2Attractor generates the Jason–Rampe II attractor and accumulates hit counts into the provided image's R buffer.
// 
// It iterates the attractor using parameters A, B, C, D from params, mapping each point to pixel coordinates using params.Scale,
// params.XOffset and params.YOffset. The first 100 iterations are used to settle and are not recorded; thereafter each pixel's hit
// count in image.R is incremented up to a cap of 200. The iteration count is taken from params.Maxiter and the generator starts
// from the initial point (0.1, 0.0). After accumulation the function calls image.RImage2IImage() to finalize the image.
func CalcJasonRampe2Attractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.1
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := jason_rampe_2(x, y, params.A, params.B, params.C, params.D)
		xi := int(float64(width)/2 + params.Scale*xn + params.XOffset)
		yi := int(float64(height)/2 + params.Scale*yn + params.YOffset)
		if i > settleDownPoints {
			if xi >= 0 && yi >= 0 && xi < int(width) && yi < int(height) {
				hits := float64(image.R[yi][xi])
				if hits < maxHits {
					image.R[yi][xi] += 1.0
				}
			}
		}
		x, y = xn, yn
	}
	image.RImage2IImage()
}