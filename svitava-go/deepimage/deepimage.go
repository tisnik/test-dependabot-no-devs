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

func New(width uint, height uint) (Image, error) {
	resolution, err := NewResolution(width, height)

	if err != nil {
		return Image{}, err
	}

	return Image{
		Resolution: resolution,
		Z:          NewZImage(resolution),
		R:          NewRImage(resolution),
		I:          NewIImage(resolution),
	}, nil
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

func (image *Image) RImage2IImage() {
	r := image.Resolution
	width := r.Width
	height := r.Height

	min, max := image.R.minMax(width, height)
	k := 255.0 / (max - min)

	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			f := float64(image.R[y][x])
			f -= min
			f *= k
			if f > 255.0 {
				f = 255
			}
			i := int(f) & 255
			image.Z[y][x] = ZPixel(complex(float32(x), float32(y)))
			image.I[y][x] = IPixel(i)
		}
	}
}

func (image *Image) RImage2IImageWithFactor(maxFactor float64) {
	r := image.Resolution
	width := r.Width
	height := r.Height

	min, max := image.R.minMax(width, height)
	max *= maxFactor
	k := 255.0 / (max - min)

	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			f := float64(image.R[y][x])
			f -= min
			f *= k
			if f > 255.0 {
				f = 255
			}
			i := int(f) & 255
			image.Z[y][x] = ZPixel(complex(float32(x), float32(y)))
			image.I[y][x] = IPixel(i)
		}
	}
}
