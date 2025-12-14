//
//  (C) Copyright 2024  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package deepimage

import (
	"math"
)

// RImage is representation of raster image consisting of RPixels
type RImage [][]RPixel

// NewRImage constructs new instance of RImage
func NewRImage(resolution Resolution) RImage {
	rimage := make([][]RPixel, resolution.Height)
	for y := uint(0); y < resolution.Height; y++ {
		rimage[y] = make([]RPixel, resolution.Width)
	}
	return rimage
}

func (image *RImage) minMax(width, height uint) (float64, float64) {
	min := float64(math.Inf(1))
	max := float64(math.Inf(-1))

	for j := range height {
		for i := range width {
			z := float64((*image)[j][i])
			if max < z {
				max = z
			}
			if min > z {
				min = z
			}
		}
	}
	return min, max
}
