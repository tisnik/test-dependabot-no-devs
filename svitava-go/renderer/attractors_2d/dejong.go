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

// de_jong computes the next coordinates of the De Jong attractor for a point (x, y) using parameters a, b, c, and d.
// It returns xn and yn where xn = sin(a*y) - cos(b*x) and yn = sin(c*x) - cos(d*y).
func de_jong(x, y, a, b, c, d float64) (float64, float64) {
	xn := math.Sin(a*y) - math.Cos(b*x)
	yn := math.Sin(c*x) - math.Cos(d*y)
	return xn, yn
}

// CalcDeJongAttractor iterates the De Jong attractor using the parameters in params
// and accumulates point hit counts into the provided image.
//
// It runs up to params.Maxiter starting from (0,0), maps attractor coordinates to
// pixel indices using params.Scale, params.XOffset and params.YOffset, and increments
// the image's red-channel hit counts for points that fall inside the image bounds
// (after a brief settle-down phase). The function modifies the supplied image in place
// and finalizes it by converting the accumulated red-channel data via image.RImage2IImage().
func CalcDeJongAttractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.0
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := de_jong(x, y, params.A, params.B, params.C, params.D)
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