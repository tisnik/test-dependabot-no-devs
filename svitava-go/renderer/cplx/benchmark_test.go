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

package cplx_test

import (
	"testing"

	"github.com/tisnik/svitava-go/params"
	"github.com/tisnik/svitava-go/renderer/cplx"
)

const (
	WIDTH   = 512
	HEIGHT  = 512
	MAXITER = 1000
)

func BenchmarkCalcMandelbrot(b *testing.B) {
	params := params.FractalParameter{
		Cx0:     0,
		Cy0:     0,
		Maxiter: MAXITER,
	}
func BenchmarkCalcMandelbrot(b *testing.B) {
	params := params.FractalParameter{
		Cx0:     0,
		Cy0:     0,
		Maxiter: MAXITER,
	}
	image, _ := deepimage.New(WIDTH, HEIGHT)
	for i := 0; i < b.N; i++ {
		cplx.CalcMandelbrot(params, image)
	}
}
	}
}

func BenchmarkCalcMandelbrotComplex(b *testing.B) {
	params := params.FractalParameter{
		Cx0:     0,
		Cy0:     0,
		Maxiter: MAXITER,
	}
	zimage := cplx.NewZImage(WIDTH, HEIGHT)
	for i := 0; i < b.N; i++ {
		cplx.CalcMandelbrotComplex(WIDTH, HEIGHT, params, zimage)
	}
}

func BenchmarkCalcJulia(b *testing.B) {
	params := params.FractalParameter{
		Cx0:     0.0,
		Cy0:     1.0,
		Maxiter: MAXITER,
	}
	zimage := cplx.NewZImage(WIDTH, HEIGHT)
	for i := 0; i < b.N; i++ {
		cplx.CalcJulia(WIDTH, HEIGHT, params, zimage)
	}
}
