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
	"log"

	"github.com/tisnik/svitava-go/deepimage"
	im "github.com/tisnik/svitava-go/image"
	"github.com/tisnik/svitava-go/palettes"
	"github.com/tisnik/svitava-go/params"
	"github.com/tisnik/svitava-go/renderer/attractors_2d"
	"github.com/tisnik/svitava-go/renderer/cplx"
	"github.com/tisnik/svitava-go/renderer/textures"
)

func init() {
	log.Println("Renderer: init")
}

type Renderer interface {
	RenderComplexFractal(resolution im.Resolution, params params.FractalParameter, palette palettes.Palette) image.Image
}

type SingleGoroutineRenderer struct {
}

func NewSingleGoroutineRenderer() Renderer {
	return SingleGoroutineRenderer{}
}

type fractalFunction = func(params params.FractalParameter, deepImage deepimage.Image)

func render(width uint, height uint, params params.FractalParameter, palette palettes.Palette, function fractalFunction) image.Image {
	if width == 0 || height == 0 {
		// TODO: logging
		return nil
	}
	if function == nil {
		// TODO: logging
		return nil
	}
	deepImage, err := deepimage.New(width, height)
	if err != nil {
		// TODO: logging
		return nil
	}
	function(params, deepImage)
	deepImage.ApplyPalette(palette)
	return deepImage.RGBA
}

func (r SingleGoroutineRenderer) RenderComplexFractal(
	resolution im.Resolution,
	params params.FractalParameter,
	palette palettes.Palette) image.Image {

	functions := map[string]fractalFunction{
		"mandelbrot":        cplx.CalcMandelbrotComplex,
		"julia":             cplx.CalcJulia,
		"mandelbrot_z3":     cplx.CalcMandelbrotZ3,
		"mandelbrot_z4":     cplx.CalcMandelbrotZ4,
		"mandelbrot_z2pz":   cplx.CalcMandelbrotZ2pZ,
		"mandelbrot_z2mz":   cplx.CalcMandelbrotZ2mZ,
		"mandelbrot_fn":     cplx.CalcMandelbrotFn,
		"julia_z3":          cplx.CalcJuliaZ3,
		"julia_z4":          cplx.CalcJuliaZ4,
		"julia_fn":          cplx.CalcJuliaFn,
		"barnsley_m1":       cplx.CalcBarnsleyMandelbrotM1,
		"barnsley_j1":       cplx.CalcBarnsleyJuliaJ1,
		"barnsley_m2":       cplx.CalcBarnsleyMandelbrotM2,
		"barnsley_j2":       cplx.CalcBarnsleyJuliaJ2,
		"barnsley_m3":       cplx.CalcBarnsleyMandelbrotM3,
		"barnsley_j3":       cplx.CalcBarnsleyJuliaJ3,
		"phoenix_m":         cplx.CalcPhoenixM,
		"phoenix_j":         cplx.CalcPhoenixJ,
		"lambda_m":          cplx.CalcMandelLambda,
		"lambda_j":          cplx.CalcLambda,
		"manowar_m":         cplx.CalcManowarM,
		"manowar_j":         cplx.CalcManowarJ,
		"mandelbrot_zpower": cplx.CalcZPowerMandelbrot,
		"newton":            cplx.CalcNewton,
		"circle_pattern":    textures.CalcCirclePattern,
		"plasma_pattern":    textures.CalcPlasmaPattern,
		"fm_synth":          textures.CalcFMSynth,
		"bedhead":           attractors_2d.CalcBedheadAttractor,
		"de_jong":           attractors_2d.CalcDeJongAttractor,
		"fractal_dream":     attractors_2d.CalcFractalDreamAttractor,
		"hopalong":          attractors_2d.CalcHopalongAttractor,
		"jason_rampe_1":     attractors_2d.CalcJasonRampe1Attractor,
		"jason_rampe_2":     attractors_2d.CalcJasonRampe2Attractor,
	}

	function, exists := functions[params.Type]
	if !exists {
		// Return default function
		function = cplx.CalcMandelbrotComplex
	}

	return render(resolution.Width, resolution.Height, params, palette, function)
}
