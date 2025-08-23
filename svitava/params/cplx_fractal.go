//
//  (C) Copyright 2024 - 2025  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package params

import (
	"errors"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Cplx structure contains information about all fractal parameters.
//
// None: currently, only fractals in complex plane are supported
type Cplx struct {
	Name      string  `toml:"name"`
	Type      string  `toml:"type"`
	Cx0       float64 `toml:"cx0"`
	Cy0       float64 `toml:"cy0"`
	Maxiter   uint    `toml:"maxiter"`
	Bailout   uint    `toml:"bailout"`
	Function1 string  `toml:"function1"`
	Function2 string  `toml:"function2"`
	Xmin      float64 `toml:"xmin"`
	Ymin      float64 `toml:"ymin"`
	Xmax      float64 `toml:"xmax"`
	Ymax      float64 `toml:"ymax"`
}

// Sequence of fractal parameters
type CplxParams struct {
	Parameters []Cplx `toml:"complex_fractal"`
}

// LoadCplxParameters function reads fractal parameters from external text file
func LoadCplxParameters(filename string) (map[string]Cplx, error) {
	var parameters CplxParams
	asMap := map[string]Cplx{}

	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return asMap, errors.New("Parameter file does not exist.")
	}
	if err != nil {
		return asMap, err
	}

	_, err = toml.DecodeFile(filename, &parameters)
	if err != nil {
		log.Fatal(err)
		return asMap, err
	}
	for _, parameter := range parameters.Parameters {
		asMap[parameter.Name] = parameter
	}
	return asMap, nil
}
