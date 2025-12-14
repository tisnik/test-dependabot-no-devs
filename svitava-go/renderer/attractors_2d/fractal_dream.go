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

// fractal_dream computes the next coordinates of the fractal-dream map for the given point and parameters.
// It returns xn = sin(b*y) - c*sin(b*x) and yn = sin(a*x) - d*sin(a*y).
func fractal_dream(x, y, a, b, c, d float64) (float64, float64) {
	xn := math.Sin(b*y) - c*math.Sin(b*x)
	yn := math.Sin(a*x) - d*math.Sin(a*y)
	return xn, yn
}

// image by calling image.RImage2IImage().
func CalcFractalDreamAttractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.1
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := fractal_dream(x, y, params.A, params.B, params.C, params.D)
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