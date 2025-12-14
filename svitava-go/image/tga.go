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

package image

import (
	"image"
	"log"
	"os"
)

// TGAImageWriter implements image.Writer interface, it writes TGA format
type TGAImageWriter struct{}

// WriteTGAImage writes an image represented by byte slice into file with TGA format.
func (writer TGAImageWriter) WriteImage(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	tgaHeader := []byte{
		/* TGA header structure: */
		0x00,       /* without image ID */
		0x00,       /* color map type: without palette */
		0x02,       /* uncompressed true color image */
		0x00, 0x00, /* start of color palette (it is not used) */
		0x00, 0x00, /* length of color palette (it is not used) */
		0x00,                   /* bits per palette entry */
		0x00, 0x00, 0x00, 0x00, /* image coordinates */
		0x00, 0x00, /* image width */
		0x00, 0x00, /* image height */
		0x18, /* bits per pixel = 24 */
		0x20, /* picture orientation: top-left origin */
	}

	/* image size is specified in TGA header */
	tgaHeader[12] = byte(width & 0xff)
	tgaHeader[13] = byte(width >> 8)
	tgaHeader[14] = byte(height & 0xff)
	tgaHeader[15] = byte(height >> 8)

	file.Write(tgaHeader)

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

// NewTGAImageWriter returns a new TGAImageWriter.
// The returned writer implements writing 24-bit uncompressed TGA images.
func NewTGAImageWriter() TGAImageWriter {
	return TGAImageWriter{}
}