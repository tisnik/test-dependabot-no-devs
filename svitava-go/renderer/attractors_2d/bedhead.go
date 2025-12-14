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

// bedhead computes the next point of the Bedhead attractor for input coordinates x and y
// using parameters a and b. It returns xn and yn where
// xn = sin(x*y/b)*y + cos(a*x - y) and yn = x + sin(y)/b.
func bedhead(x, y, a, b float64) (float64, float64) {
	xn := math.Sin(x*y/b)*y + math.Cos(a*x-y)
	yn := x + math.Sin(y)/b
	return xn, yn
}

// CalcBedheadAttractor renders the "Bedhead" 2D attractor into the provided image buffer.
// 
// CalcBedheadAttractor iterates the bedhead map using parameters from params and accumulates
// hit counts into image.R after an initial settling phase. The function starts from (0,0),
// runs params.Maxiter iterations, maps each computed point (xn,yn) to pixel coordinates using
// params.Scale, params.XOffset and params.YOffset, and—after 100 settling iterations—increments
// the corresponding red-channel accumulator while capping increments at 200 hits per pixel.
// At the end it converts the floating-point red-channel buffer to the image's integer form.
//
// params: fractal parameters supplying A, B, Scale, XOffset, YOffset and Maxiter.
// image: deepimage.Image whose Resolution and R buffer are updated in-place.
func CalcBedheadAttractor(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	settleDownPoints := 100
	maxHits := 200.0

	x := 0.0
	y := 0.0

	for i := range int(params.Maxiter) {
		xn, yn := bedhead(x, y, params.A, params.B)
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