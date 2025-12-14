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

package deepimage

import (
	"errors"
	"fmt"
)

// Resolution describes the image dimensions in pixels.
type Resolution struct {
	Width  uint
	Height uint
}

// NewResolution constructs a Resolution with the given width and height.
// NewResolution constructs a Resolution with the given width and height.
// It returns an error if width or height is zero or greater than 65535; on success it returns the Resolution and a nil error.
func NewResolution(width, height uint) (Resolution, error) {
	// Check for zero dimensions
	if width == 0 {
		return Resolution{}, errors.New("width cannot be zero")
	}
	if height == 0 {
		return Resolution{}, errors.New("height cannot be zero")
	}

	// Check for reasonable maximum dimensions to prevent memory issues
	const maxDimension = 65535 // 2^16 - 1, reasonable for image processing
	if width > maxDimension {
		return Resolution{}, fmt.Errorf("width %d exceeds maximum allowed dimension %d", width, maxDimension)
	}
	if height > maxDimension {
		return Resolution{}, fmt.Errorf("height %d exceeds maximum allowed dimension %d", height, maxDimension)
	}

	return Resolution{
		Width:  width,
		Height: height,
	}, nil
}