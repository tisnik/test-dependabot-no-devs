//
//  (C) Copyright 2019, 2020, 2024  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package image

import (
	"image"
	"log"
	"os"
)

// BMPImageWriter implements image.Writer interface, it writes BMP format
type BMPImageWriter struct{}

// WriteImage writes an image represented by byte slice into file with BMP format.
func (writer BMPImageWriter) WriteImage(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	bmpHeader := []byte{
		/* BMP header structure: */
		0x42, 0x4d, /* magic number */
		0x46, 0x00, 0x00, 0x00, /* size of header=70 bytes */
		0x00, 0x00, /* unused */
		0x00, 0x00, /* unused */
		0x36, 0x00, 0x00, 0x00, /* 54 bytes - offset to data */
		0x28, 0x00, 0x00, 0x00, /* 40 bytes - bytes in DIB header */
		0x00, 0x00, 0x00, 0x00, /* width of bitmap */
		0x00, 0x00, 0x00, 0x00, /* height of bitmap */
		0x01, 0x0, /* 1 pixel plane */
		0x18, 0x00, /* 24 bpp */
		0x00, 0x00, 0x00, 0x00, /* no compression */
		0x00, 0x00, 0x00, 0x00, /* size of pixel array */
		0x13, 0x0b, 0x00, 0x00, /* 2835 pixels/meter */
		0x13, 0x0b, 0x00, 0x00, /* 2835 pixels/meter */
		0x00, 0x00, 0x00, 0x00, /* color palette */
		0x00, 0x00, 0x00, 0x00, /* important colors */
	}

	bmpHeader[18] = byte(width & 0xff)
	bmpHeader[19] = byte(width >> 8)
	bmpHeader[20] = byte(width >> 16)
	bmpHeader[21] = byte(width >> 24)
	bmpHeader[22] = byte(height & 0xff)
	bmpHeader[23] = byte(height >> 8)
	bmpHeader[24] = byte(height >> 16)
	bmpHeader[25] = byte(height >> 24)

	file.Write(bmpHeader)

	for y := range height {
		for x := range width {
			r, g, b, _ := img.At(x, y).RGBA()
			// swap RGB
			color := []byte{byte(b >> 8), byte(g >> 8), byte(r >> 8)}
			file.Write(color)
		}
	}
	// no error
	return nil
}

// NewBMPImageWriter is a constructor for BMP image writer
func NewBMPImageWriter() BMPImageWriter {
	return BMPImageWriter{}
}
