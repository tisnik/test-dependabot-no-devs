# Svitava

[![Go Report Card](https://goreportcard.com/badge/github.com/tisnik/svitava-go)](https://goreportcard.com/report/github.com/tisnik/svitava-go)

Concurrent fractal renderer written in the Go programming language

<!-- vim-markdown-toc GFM -->

* [Installation](#installation)
    * [Prerequisities](#prerequisities)
* [Starting the application](#starting-the-application)
* [Configuration](#configuration)
    * [Main configuration file](#main-configuration-file)
    * [Fractal parameters files](#fractal-parameters-files)
* [Developing](#developing)
* [Contribution](#contribution)

<!-- vim-markdown-toc -->

## Installation

* It is needed to have Go tooling installed. It can be Go 1.20 or newer. Please look at https://go.dev/dl/ page with options how to download and install the required Go tooling.

* To build the application, use the following command:
```bash
go build
```

### Prerequisities

## Starting the application

## Configuration

### Main configuration file

Main configuration file is named `config.toml`. It currently has the following format:

```toml
[server]
address = ":8080"

[logging]
debug = true

[rendering]
image_format = "png"
```

### Fractal parameters files

```toml
[[complex_fractal]]
name = "Classic Mandelbrot set"
type = "mandelbrot"
cx0 = 0.0
cy0 = 0.0
maxiter = 1000
bailout = 4
xmin = -2.0
xmax = 1.0
ymin = -1.5
ymax = 1.5

[[complex_fractal]]
name = "Mandelbrot set z=z^3+c"
type = "mandelbrot_z3"
cx0 = 0.0
cy0 = 0.0
maxiter = 1000

...
...
...
```

## Developing

## Contribution

Please look into document [CONTRIBUTING.md](CONTRIBUTING.md) that contains all information about how to contribute to this project.
