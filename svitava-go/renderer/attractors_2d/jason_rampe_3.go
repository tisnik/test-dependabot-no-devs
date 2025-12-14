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

// jason_rampe_3 computes the next point of the Jason-Rampe-3 attractor for the given coordinates and parameters.
// It returns xn, yn where xn = sin(b*y) + c*cos(b*x) and yn = cos(a*x) + d*sin(a*y).
func jason_rampe_3(x, y, a, b, c, d float64) (float64, float64) {
	xn := math.Sin(b*y) + c*math.Cos(b*x)
	yn := math.Cos(a*x) + d*math.Sin(a*y)
	return xn, yn
}

// CalcJasonRampe3Attractor computes the Jason‑Rampe‑3 attractor using the supplied fractal parameters
// and accumulates visit counts into the image's red channel, then converts that accumulation to the
// image's final representation.
//
// It iterates up to params.Maxiter starting from (0.1, 0.0), maps attractor coordinates to pixel
// indices using params.Scale, params.XOffset and params.YOffset around the image center, and ignores
// the first 100 points as a settling period. After settling, each mapped pixel's red channel is
// incremented for each visit up to a per-pixel cap of 200 hits. Once iteration completes,
// image.RImage2IImage() is called to produce the resulting image.
func CalcJasonRampe3Attractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.1
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := jason_rampe_3(x, y, params.A, params.B, params.C, params.D)
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