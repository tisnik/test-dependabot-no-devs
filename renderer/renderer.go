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

func init() {
	println("Init")
}

type Renderer interface {
	RenderComplexFractal(resolution im.Resolution, params params.Cplx, palette palettes.Palette) image.Image
}

type SingleGoroutineRenderer struct {
}

func NewSingleGoroutineRenderer() Renderer {
	return SingleGoroutineRenderer{}
}

type fractalFunction = func(params params.Cplx, deepImage deepimage.Image)

func render(width uint, height uint, params params.Cplx, palette palettes.Palette, function fractalFunction) image.Image {
	deepImage := deepimage.New(width, height)
	function(params, deepImage)
	deepImage.ApplyPalette(palette)
	return deepImage.RGBA
}

func (r SingleGoroutineRenderer) RenderComplexFractal(resolution im.Resolution, params params.Cplx, palette palettes.Palette) image.Image {
	functions := map[string]fractalFunction{}
	// TODO: change to params.Type!!!
	functions["Classic Mandelbrot set"] = cplx.CalcMandelbrotComplex
	functions["Mandelbrot set z=z^3+c"] = cplx.CalcMandelbrotZ3
	functions["Mandelbrot set z=z^4+c"] = cplx.CalcMandelbrotZ4
	functions["Phoenix set, Mandelbrot variant"] = cplx.CalcPhoenixM
	functions["Phoenix set, Julia variant"] = cplx.CalcPhoenixJ
	functions["Lambda, Mandelbrot variant"] = cplx.CalcMandelLambda
	functions["Lambda, Julia variant"] = cplx.CalcLambda
	functions["Manowar, Mandelbrot variant"] = cplx.CalcManowarM

	return render(resolution.Width, resolution.Height, params, palette, functions[params.Name])
}

// RenderMandelbrotFractal renders a classic Mandelbrot fractal into provided Image.
func RenderMandelbrotFractal(width uint, height uint, pcx float64, pcy float64, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0,
		Cy0:     0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcMandelbrotComplex)
}

// RenderJuliaFractal renders a classic Julia fractal into provided Image.
func RenderJuliaFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcJulia)
}

// RenderBarnsleyFractalM1 renders a classic Barnsley fractal M1 into provided Image.
func RenderBarnsleyFractalM1(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM1)
}

// RenderBarnsleyFractalM2 renders a classic Barnsley fractal M2 into provided Image.
func RenderBarnsleyFractalM2(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM2)
}

// RenderBarnsleyFractalM3 renders a classic Barnsley fractal M3 into provided Image.
func RenderBarnsleyFractalM3(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyM3)
}

// RenderBarnsleyFractalJ1 renders a classic Barnsley fractal J1 into provided Image.
func RenderBarnsleyFractalJ1(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.48,
		Cy0:     -1.32,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ1)
}

// RenderBarnsleyFractalJ2 renders a classic Barnsley fractal J2 into provided Image.
func RenderBarnsleyFractalJ2(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.5,
		Cy0:     1.2,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ2)
}

// RenderBarnsleyFractalJ3 renders a classic Barnsley fractal J3 into provided Image.
func RenderBarnsleyFractalJ3(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     0.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcBarnsleyJ3)
}

// RenderMagnet renders a classic Magnet fractal into provided Image.
func RenderMagnetFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	/*params := params.Cplx{
		Cx0:     1.1,
		Cy0:     1.1,
		Maxiter: 1000,
	}*/
	params := params.Cplx{
		Cx0:     0.0,
		Cy0:     0.0,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcMagnet)
}

// RenderMagnet renders a classic Magnet Julia fractal into provided Image.
func RenderMagnetJuliaFractal(width uint, height uint, maxiter uint, palette palettes.Palette) image.Image {
	params := params.Cplx{
		Cx0:     0.5,
		Cy0:     -1.5,
		Maxiter: 1000,
	}
	return render(width, height, params, palette, cplx.CalcMagnetJulia)
}
