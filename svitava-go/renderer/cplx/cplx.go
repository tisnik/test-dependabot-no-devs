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

// getSteps computes the horizontal and vertical coordinate step sizes per image pixel.
// The first returned value is the X step ((Xmax - Xmin) / image width) and the second is the Y step ((Ymax - Ymin) / image height).
func getSteps(
	params params.FractalParameter,
	image deepimage.Image) (float64, float64) {
	stepX := float64(params.Xmax-params.Xmin) / float64(image.Resolution.Width)
	stepY := float64(params.Ymax-params.Ymin) / float64(image.Resolution.Height)
	return stepX, stepY
}

// calcIndex computes a palette index from an iteration count using the parameterized shift and slope.
// It applies params.Palette.Shift + int(i)*params.Palette.Slope, clamps negative results to 0, and returns the result as a uint.
func calcIndex(params params.FractalParameter, i uint) uint {
	index := params.Palette.Shift + int(i)*params.Palette.Slope
	if index < 0 {
		return 0
	}
	return uint(index)
}