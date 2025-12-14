//
//  (C) Copyright 2025  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package textures

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// getSteps computes the horizontal and vertical step sizes that map image pixels to the fractal's coordinate ranges.
// It returns stepX and stepY where stepX = (params.Xmax - params.Xmin) / image.Resolution.Width and
// stepY = (params.Ymax - params.Ymin) / image.Resolution.Height.
func getSteps(
	params params.FractalParameter,
	image deepimage.Image) (float64, float64) {
	stepX := float64(params.Xmax-params.Xmin) / float64(image.Resolution.Width)
	stepY := float64(params.Ymax-params.Ymin) / float64(image.Resolution.Height)
	return stepX, stepY
}