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

package cplx

import (
	"math/cmplx"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcMandelbrotFn computes a Mandelbrot-like fractal over the provided image and writes per-pixel complex results and iteration counts into image.Z and image.I.
// For each pixel it maps (x,y) to the complex plane, iterates z = c * sin(z) until the squared magnitude exceeds params.Bailout or params.Maxiter is reached, then stores the final z and the iteration count.
func CalcMandelbrotFn(
	params params.FractalParameter,
	image deepimage.Image) {

	stepX, stepY := getSteps(params, image)

	var cy float64 = params.Ymin
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = params.Xmin
		for x := uint(0); x < image.Resolution.Width; x++ {
			c := complex(cx, cy)
			z := c
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > float64(params.Bailout) {
					break
				}
				z = c * cmplx.Sin(z)
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += stepX
		}
		cy += stepY
	}
}