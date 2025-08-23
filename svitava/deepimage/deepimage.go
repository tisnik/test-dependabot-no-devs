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
	"image"

	"github.com/tisnik/svitava-go/palettes"
)

type Image struct {
	Resolution Resolution
	Z          ZImage
	R          RImage
	I          IImage
	RGBA       *image.NRGBA
}

func New(width uint, height uint) Image {
	resolution := NewResolution(width, height)

	return Image{
		Resolution: resolution,
		Z:          NewZImage(resolution),
		R:          NewRImage(resolution),
		I:          NewIImage(resolution),
	}
}

func (i *Image) ApplyPalette(palette palettes.Palette) {
	r := i.Resolution
	i.RGBA = image.NewNRGBA(image.Rect(0, 0, int(r.Width), int(r.Height)))

	for y := 0; y < int(r.Height); y++ {
		offset := i.RGBA.PixOffset(0, y)
		for x := uint(0); x < r.Width; x++ {
			index := byte(i.I[y][x])
			i.RGBA.Pix[offset] = palette[index][0]
			offset++
			i.RGBA.Pix[offset] = palette[index][1]
			offset++
			i.RGBA.Pix[offset] = palette[index][2]
			offset++
			i.RGBA.Pix[offset] = 0xff
			offset++
		}
	}
}
