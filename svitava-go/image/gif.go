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
	"image/gif"
	"os"
)

// GIFImageWriter implements image.Writer interface, it writes GIF format
type GIFImageWriter struct{}

// WriteGIFImage writes an image represented by standard image.Image structure into file with GIF format.
func (writer GIFImageWriter) WriteImage(filename string, img image.Image) error {
	outfile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	return gif.Encode(outfile, img, nil)
}

// NewGIFImageWriter returns a GIFImageWriter configured for writing GIF images.
func NewGIFImageWriter() GIFImageWriter {
	return GIFImageWriter{}
}