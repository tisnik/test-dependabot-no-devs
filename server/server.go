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

package server

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strings"

	"github.com/tisnik/svitava-go/image"
	"github.com/tisnik/svitava-go/palettes"
	"github.com/tisnik/svitava-go/params"
	"github.com/tisnik/svitava-go/renderer"
)

// Server interface can be satisfied by any structure that implements Serve()
// method
type Server interface {
	Serve()
}

// HTTPServer structure that satisfy Server interface
type HTTPServer struct {
	port     uint
	renderer renderer.Renderer
}

// NewHTTPServer constructs new instance of HTTP server
func NewHTTPServer(port uint, renderer renderer.Renderer) Server {
	return HTTPServer{
		port:     port,
		renderer: renderer,
	}
}

func (s HTTPServer) indexPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web-content/index.html")
}

func (s HTTPServer) newFractalPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web-content/new_fractal.html")
}

func (s HTTPServer) galleryPageHandler(w http.ResponseWriter, r *http.Request) {
}

func (s HTTPServer) settingsPageHandler(w http.ResponseWriter, r *http.Request) {
}

func (s HTTPServer) staticImageHandler(w http.ResponseWriter, r *http.Request) {
	imageName := r.URL.String()
	fileName := strings.TrimPrefix(imageName, "/image/")
	fullPath := "web-content/images/" + fileName
	http.ServeFile(w, r, fullPath)
}

func (s HTTPServer) mandelbrotPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web-content/mandelbrot.html")
}

func (s HTTPServer) complexFractalsPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web-content/complex.html")
}

func (s HTTPServer) staticIconHandler(w http.ResponseWriter, r *http.Request) {
	iconName := r.URL.String()
	fileName := strings.TrimPrefix(iconName, "/icons/")
	fullPath := "web-content/icons/" + fileName
	http.ServeFile(w, r, fullPath)
}

func (s HTTPServer) styleSheetHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web-content/svitava.css")
}

func (s HTTPServer) fractalImageHandler(w http.ResponseWriter, r *http.Request) {
	fractalName, err := parseStringQueryParameter(r, "fractal", "Classic Mandelbrot set")
	if err != nil {
		log.Fatal(err)
	}

	paletteName, err := parseStringQueryParameter(r, "palette", "mandmap.map")
	if err != nil {
		log.Fatal(err)
	}

	width, err := parseUintQueryParameter(r, "width", 128)
	if err != nil {
		log.Fatal(err)
	}

	height, err := parseUintQueryParameter(r, "height", 128)
	if err != nil {
		log.Fatal(err)
	}

	resolution := image.Resolution{
		Width:  uint(width),
		Height: uint(height),
	}
	fmt.Println(fractalName, paletteName, resolution)
	palette, _ := palettes.LoadTextRGBPalette("data/" + paletteName + ".map")
	parameters, _ := params.LoadCplxParameters("data/complex_fractals.toml")
	img := s.renderer.RenderComplexFractal(resolution, parameters["Classic Mandelbrot set"], palette)
	png.Encode(w, img)
}

// Serve method starts HTTP server that provides all static and dynamic data
func (s HTTPServer) Serve() {
	log.Printf("Starting server on port %d", s.port)
	/* http.Handle("/", http.FileServer(http.Dir("web-content/"))) */

	http.HandleFunc("/", s.indexPageHandler)
	http.HandleFunc("/new-fractal", s.newFractalPageHandler)
	http.HandleFunc("/gallery", s.galleryPageHandler)
	http.HandleFunc("/settings", s.settingsPageHandler)
	http.HandleFunc("/svitava.css", s.styleSheetHandler)
	http.HandleFunc("/icons/{name}", s.staticIconHandler)
	http.HandleFunc("/image/new_fractal/{path}", s.staticImageHandler)
	http.HandleFunc("/mandelbrot", s.mandelbrotPageHandler)
	http.HandleFunc("/complex", s.complexFractalsPageHandler)
	//http.HandleFunc("/image/main/{type}", s.staticImageHandler)
	http.HandleFunc("/render", s.fractalImageHandler)

	//imageServer := http.FileServer(http.Dir("web-content/images/"))
	//http.Handle("/image/", http.StripPrefix("/image", imageServer))

	http.ListenAndServe(":8080", nil)
}
