//
//  (C) Copyright 2019, 2020, 2021, 2022, 2023, 2024  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/tisnik/svitava-go/configuration"
	"github.com/tisnik/svitava-go/image"
	"github.com/tisnik/svitava-go/palettes"
	"github.com/tisnik/svitava-go/renderer"
	"github.com/tisnik/svitava-go/server"

	"github.com/tisnik/svitava-go/params"
)

const (
	CONFIG_FILE_NAME = "config.toml"
)

func runInDemoMode() {
	log.Println("Starting demo mode: render all fractals available")

	palette, err := palettes.LoadTextRGBPalette("data/blues.map")
	log.Println("Color palette loaded")

	resolution := image.Resolution{
		Width:  512,
		Height: 512,
	}

	r := renderer.NewSingleGoroutineRenderer()

	parameters, err := params.LoadFractalParameters("data/complex_fractals.toml")
	log.Printf("Fractal configuration:  %v  %v", parameters, err)

	var writer image.Writer
	writer = image.NewBMPImageWriter()
	log.Println("BMP image writer initialized")

	fractals := []string{
		"Classic Mandelbrot set",
		"Classic Julia set",
		"Mandelbrot set z=z^3+c",
		"Mandelbrot set z=z^4+c",
		"Mandelbrot set z=z^2-z+c",
		"Phoenix set, Mandelbrot variant",
		"Phoenix set, Julia variant",
		"Lambda, Mandelbrot variant",
		"Lambda, Julia variant",
		"Manowar, Mandelbrot variant",
		"Manowar, Julia variant",
	}

	for _, fractal := range fractals {
		log.Println("Rendering", fractal, "started")
		t1 := time.Now()
		img := r.RenderComplexFractal(resolution, parameters[fractal], palette)
		writer.WriteImage(fractal+".bmp", img)
		t2 := time.Now()
		log.Println("Rendering", fractal, "finished in", t2.Sub(t1))
	}
}

func runInServerMode(port uint) {
	log.Println("Starting server")
	r := renderer.NewSingleGoroutineRenderer()
	server := server.NewHTTPServer(port, r)
	server.Serve()
}

func listAllFractals() {
	parameters, _ := params.LoadFractalParameters("data/complex_fractals.toml")

	names := make([]string, len(parameters))
	i := 0
	for name := range parameters {
		names[i] = name
		i++
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

func main() {
	var width uint
	var height uint
	var aa bool
	var startServer bool
	var startTUI bool
	var execute string
	var port uint
	var demoMode bool
	var fractal string
	var listFractals bool

	configuration, err := configuration.LoadConfiguration(CONFIG_FILE_NAME)
	if err != nil {
		println("Unable to load configuration")
		os.Exit(1)
	}
	log.Println("Configuration:", configuration)

	flag.UintVar(&width, "w", 0, "image width (shorthand)")
	flag.UintVar(&width, "width", 0, "image width")

	flag.UintVar(&height, "h", 0, "image height (shorthand)")
	flag.UintVar(&height, "height", 0, "image height")

	flag.BoolVar(&aa, "a", false, "enable antialiasing (shorthand)")
	flag.BoolVar(&aa, "antialias", false, "enable antialiasing")

	flag.BoolVar(&startTUI, "t", false, "start with text user interface (shorthand)")
	flag.BoolVar(&startTUI, "tui", false, "start with text user interface")

	flag.BoolVar(&listFractals, "l", false, "list names of all fractals that can be rendered (shorthand)")
	flag.BoolVar(&listFractals, "list", false, "list names of all fractals that can be rendered")

	flag.StringVar(&fractal, "f", "", "name of fractal to be rendered (shorthand)")
	flag.StringVar(&fractal, "fractal", "", "name of fractal to be rendered")

	flag.StringVar(&execute, "e", "", "execute given script with rendering commands (shorthand)")
	flag.StringVar(&execute, "exec", "", "execute given script with rendering commands")
	flag.StringVar(&execute, "execute", "", "execute given script with rendering commands")

	flag.BoolVar(&startServer, "s", false, "start in server mode (shorthand)")
	flag.BoolVar(&startServer, "server", false, "start in server mode")

	flag.UintVar(&port, "p", 8080, "port for the server (shorthand)")
	flag.UintVar(&port, "port", 8080, "port for the server")

	flag.BoolVar(&demoMode, "d", false, "start in demo mode (render all fractals)")
	flag.BoolVar(&demoMode, "demo", false, "start in demo mode (render all fractals)")

	flag.Parse()

	if startServer {
		runInServerMode(port)
		return
	}

	if demoMode {
		runInDemoMode()
		return
	}

	if listFractals {
		listAllFractals()
		return
	}

	fmt.Println("Please choose server mode or demo mode")
}
