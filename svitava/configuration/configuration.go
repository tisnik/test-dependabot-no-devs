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

package configuration

import (
	"errors"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Configuration structure
type Configuration struct {
	ServerConfiguration    ServerConfiguration    `toml:"server"`
	LoggingConfiguration   LoggingConfiguration   `toml:"logging"`
	RenderingConfiguration RenderingConfiguration `toml:"rendering"`
}

// Server configuration
type ServerConfiguration struct {
	Address string `toml:"address"`
}

// Logging configuration
type LoggingConfiguration struct {
	Debug bool `toml:"debug"`
}

// Fractal rendering configuration
type RenderingConfiguration struct {
	ImageFormat string `toml:"image_format"`
	BinaryPPM   bool   `toml:"binary_bpm"`
}

// LoadConfiguration function reads configuration stored in TOML format
func LoadConfiguration(configFileName string) (Configuration, error) {
	var configuration Configuration

	_, err := os.Stat(configFileName)

	if os.IsNotExist(err) {
		return configuration, errors.New("Config file does not exist.")
	}
	if err != nil {
		return configuration, err
	}

	_, err = toml.DecodeFile(configFileName, &configuration)
	if err != nil {
		log.Fatal(err)
		return configuration, err
	}
	return configuration, nil
}
