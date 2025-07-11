//
//  (C) Copyright 2019 - 2025  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package renderer

import (
	"image"

	"github.com/tisnik/svitava-go/deepimage"
	im "github.com/tisnik/svitava-go/image"
	"github.com/tisnik/svitava-go/palettes"
	"github.com/tisnik/svitava-go/params"
	"github.com/tisnik/svitava-go/renderer/cplx"
)

// init prints "Init" when the package is initialized.
func init() {
	println("Init")
}

type Renderer interface {
	RenderComplexFractal(resolution im.Resolution, params params.Cplx, palette palettes.Palette) image.Image
}

type SingleGoroutineRenderer struct {
}

// NewSingleGoroutineRenderer returns a Renderer that renders fractals using a single goroutine.
func NewSingleGoroutineRenderer() Renderer {
	return SingleGoroutineRenderer{}
}

type fractalFunction = func(params params.Cplx, deepImage deepimage.Image)

// render generates a fractal image of the specified dimensions using the provided fractal parameters, palette, and fractal calculation function.
// It returns the resulting image, or nil if the dimensions are zero or the fractal function is nil.
func render(width uint, height uint, params params.Cplx, palette palettes.Palette, function fractalFunction) image.Image {
	if width == 0 || height == 0 {
		// TODO: logging
		return nil
	}
	if function == nil {
		// TODO: logging
		return nil
	}
	deepImage := deepimage.New(width, height)
	function(params, deepImage)
	deepImage.ApplyPalette(palette)
	return deepImage.RGBA
}

func (r SingleGoroutineRenderer) RenderComplexFractal(resolution im.Resolution, params params.Cplx, palette palettes.Palette) image.Image {
	functions := map[string]fractalFunction{
		"Classic Mandelbrot set":          cplx.CalcMandelbrotComplex,
		"Mandelbrot set z=z^3+c":          cplx.CalcMandelbrotZ3,
		"Mandelbrot set z=z^4+c":          cplx.CalcMandelbrotZ4,
		"Phoenix set, Mandelbrot variant": cplx.CalcPhoenixM,
		"Phoenix set, Julia variant":      cplx.CalcPhoenixJ,
		"Lambda, Mandelbrot variant":      cplx.CalcMandelLambda,
		"Lambda, Julia variant":           cplx.CalcLambda,
		"Manowar, Mandelbrot variant":     cplx.CalcManowarM,
	}

	function, exists := functions[params.Name]
	if !exists {
		// Return default function
		function = cplx.CalcMandelbrotComplex
	}

	return render(resolution.Width, resolution.Height, params, palette, function)
}

// RenderMandelbrotFractal generates an image of the classic Mandelbrot fractal using the specified dimensions, maximum iterations, and color palette.
func RenderMandelbrotFractal(width uint, height uint, pcx float64, pcy float64, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0,
		Cy0:     0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcMandelbrotComplex)
}

// RenderJuliaFractal generates an image of the classic Julia fractal with preset parameters.
//
// The fractal is rendered using the specified image dimensions, maximum iteration count, and color palette.
// The Julia set is centered at (0.0, 1.0) in the complex plane.
func RenderJuliaFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcJulia)
}

// RenderBarnsleyFractalM1 generates an image of the Barnsley fractal variant M1 using preset parameters.
// The fractal is rendered with the specified dimensions, maximum iterations, and color palette.
func RenderBarnsleyFractalM1(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM1)
}

// RenderBarnsleyFractalM2 generates an image of the Barnsley fractal variant M2 using preset parameters and the specified palette.
func RenderBarnsleyFractalM2(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM2)
}

// RenderBarnsleyFractalM3 generates an image of the Barnsley fractal variant M3 using preset parameters and the specified palette.
func RenderBarnsleyFractalM3(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM3)
}

// RenderBarnsleyFractalJ1 generates an image of the Barnsley J1 fractal using preset parameters and the specified palette.
func RenderBarnsleyFractalJ1(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.48,
		Cy0:     -1.32,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ1)
}

// RenderBarnsleyFractalJ2 generates an image of the Barnsley J2 fractal using preset parameters and the specified color palette.
func RenderBarnsleyFractalJ2(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.5,
		Cy0:     1.2,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ2)
}

// RenderBarnsleyFractalJ3 generates an image of the Barnsley J3 fractal using the specified dimensions, iteration limit, and color palette.
func RenderBarnsleyFractalJ3(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     0.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ3)
}

// RenderMagnetFractal generates an image of the classic Magnet fractal using the specified dimensions, maximum iterations, and color palette.
func RenderMagnetFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	/*params := params.Cplx{
		Cx0:     1.1,
		Cy0:     1.1,
		Maxiter: maxiter,
	}*/
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     0.0,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcMagnet)
}

// RenderMagnetJuliaFractal generates a Magnet Julia fractal image with the specified dimensions, maximum iterations, and color palette.
func RenderMagnetJuliaFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.5,
		Cy0:     -1.5,
		Maxiter: maxiter,
	}
	return render(width, height, params, palette, cplx.CalcMagnetJulia)
}
