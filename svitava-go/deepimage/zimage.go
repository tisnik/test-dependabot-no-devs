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

// ZImage is representation of raster image consisting of ZPixels
type ZImage [][]ZPixel

// NewZImage creates a ZImage with the given resolution.
// The returned ZImage has height equal to resolution.Height and each row has length resolution.Width; pixels are the zero value of ZPixel.
func NewZImage(resolution Resolution) ZImage {
	zimage := make([][]ZPixel, resolution.Height)
	for y := uint(0); y < resolution.Height; y++ {
		zimage[y] = make([]ZPixel, resolution.Width)
	}
	return zimage
}