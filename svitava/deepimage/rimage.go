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

// RImage is representation of raster image consisting of RPixels
type RImage [][]RPixel

// NewRImage constructs new instance of RImage
func NewRImage(resolution Resolution) RImage {
	rimage := make([][]RPixel, resolution.Height)
	for i := uint(0); i < resolution.Height; i++ {
		rimage[i] = make([]RPixel, resolution.Width)
	}
	return rimage
}
