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

// NewZImage constructs new instance of ZImage
func NewZImage(resolution Resolution) ZImage {
	zimage := make([][]ZPixel, resolution.Height)
	for i := uint(0); i < resolution.Height; i++ {
		zimage[i] = make([]ZPixel, resolution.Width)
	}
	return zimage
}
