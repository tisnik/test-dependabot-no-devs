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

// IImage is representation of raster image consisting of IPixels
type IImage [][]IPixel

// NewIImage constructs new instance of ZImage
func NewIImage(resolution Resolution) IImage {
	iimage := make([][]IPixel, resolution.Height)
	for y := uint(0); y < resolution.Height; y++ {
		iimage[y] = make([]IPixel, resolution.Width)
	}
	return iimage
}
