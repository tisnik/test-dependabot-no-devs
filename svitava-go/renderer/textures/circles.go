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

package textures

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

func CalcCirclePattern(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	y1 := params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		x1 := params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			x2 := x1 * x1
			y2 := y1 * y1
			i := (int)(x2+y2) & 255
			image.Z[y][x] = deepimage.ZPixel(complex(x2, y2))
			image.I[y][x] = deepimage.IPixel(i)
			x1 += stepX
		}
		y1 += stepY
	}
}
