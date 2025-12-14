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

package image

import (
	"image"
	"image/jpeg"
	"os"
)

// JPEGImageWriter implements image.Writer interface, it writes JPEG format
type JPEGImageWriter struct{}

// WriteImage writes an image represented by standard image.Image structure into file with JPEG format.
func (writer JPEGImageWriter) WriteImage(filename string, img image.Image) error {
	outfile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	return jpeg.Encode(outfile, img, nil)
}

// NewJPEGImageWriter is a constructor for JPEG image writer
func NewJPEGImageWriter() JPEGImageWriter {
	return JPEGImageWriter{}
}
