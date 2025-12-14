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
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Palette struct {
	Name  string `toml:"name"`
	Shift int    `toml:"shift"`
	Slope int    `toml:"slope"`
}

// FractalParameter structure contains information about all fractal parameters.
type FractalParameter struct {
	Name      string  `toml:"name"`
	Type      string  `toml:"type"`
	Class     string  `toml:"class"`
	Cx0       float64 `toml:"cx0"`
	Cy0       float64 `toml:"cy0"`
	Palette   Palette `toml:"palette"`
	Maxiter   uint    `toml:"maxiter"`
	Bailout   uint    `toml:"bailout"`
	Function1 string  `toml:"function1"`
	Function2 string  `toml:"function2"`
	Xmin      float64 `toml:"xmin"`
	Ymin      float64 `toml:"ymin"`
	Xmax      float64 `toml:"xmax"`
	Ymax      float64 `toml:"ymax"`
	A         float64 `toml:"A"`
	B         float64 `toml:"B"`
	C         float64 `toml:"C"`
	D         float64 `toml:"D"`
	Scale     float64 `toml:"scale"`
	XOffset   float64 `toml:"x_offset"`
	YOffset   float64 `toml:"y_offset"`
}

// Sequence of fractal parameters
type FractalParameters struct {
	Parameters []FractalParameter `toml:"fractal"`
}

// LoadFractalParameters function reads fractal parameters from external text file
func LoadFractalParameters(filename string) (map[string]FractalParameter, error) {
	var parameters FractalParameters
	asMap := map[string]FractalParameter{}

	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return asMap, errors.New("Parameter file does not exist.")
	}
	if err != nil {
		log.Fatal(err)
		return asMap, err
	}

	_, err = toml.DecodeFile(filename, &parameters)
	if err != nil {
		log.Fatal(err)
		return asMap, err
	}
	for _, parameter := range parameters.Parameters {
		if _, exists := asMap[parameter.Name]; exists {
			return asMap, fmt.Errorf(
				"duplicate parameter name %q in %s",
				parameter.Name, filename)
		}
		if parameter.Palette.Name == "" {
			parameter.Palette.Slope = 1
		}
		asMap[parameter.Name] = parameter
	}
	return asMap, nil
}
