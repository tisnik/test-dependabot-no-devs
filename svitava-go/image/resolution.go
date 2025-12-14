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

type Resolution struct {
	Width  uint
	Height uint
}

// NewResolution returns a Resolution with the specified width and height.
func NewResolution(width, height uint) Resolution {
	return Resolution{
		Width:  width,
		Height: height,
	}
}