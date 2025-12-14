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

// hopalong computes the next point of the Hopalong attractor given the current
// coordinates x, y and parameters a, b, c. It returns xn = y - sign(x)*sqrt(|b*x - c|)
// and yn = a - x.
func hopalong(x, y, a, b, c float64) (float64, float64) {
	xn := y - sign(x)*math.Sqrt(math.Abs(b*x-c))
	yn := a - x
	return xn, yn
}

// CalcHopalongAttractor generates a Hopalong attractor into the provided image using the given fractal parameters.
// It iterates the Hopalong map up to params.Maxiter starting from (0,0), maps each point to image pixel coordinates using
// params.Scale, params.XOffset and params.YOffset, and accumulates hit counts in image.R after an initial settling period.
// Hit counts are capped to prevent unbounded accumulation; after iteration the function converts the accumulated R-channel data
// into the image's integer representation via image.RImage2IImage().
func CalcHopalongAttractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.0
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := hopalong(x, y, params.A, params.B, params.C)
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