//
//  (C) Copyright 2019, 2020  Pavel Tisnovsky
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
	"image/png"
	"os"
)

// PNGImageWriter implements image.Writer interface, it writes PNG format
type PNGImageWriter struct{}

// WritePNGImage writes an image represented by standard image.Image structure into file with PNG format.
func (writer PNGImageWriter) WriteImage(filename string, img image.Image) error {
	outfile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	return png.Encode(outfile, img)
}

// NewPNGImageWriter is a constructor for PNG image writer
func NewPNGImageWriter() PNGImageWriter {
	return PNGImageWriter{}
}
