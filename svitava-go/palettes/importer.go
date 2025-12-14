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

package palettes

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// LoadBinaryRGBPalette loads an RGB palette from the binary file specified by filename.
// It returns the parsed Palette or a non-nil error if the file cannot be read or the data is invalid.
func LoadBinaryRGBPalette(filename string) (Palette, error) {
	// TODO: implementation is missing
	p := Palette{}
	return p, nil
}

// LoadBinaryRGBAPalette loads an RGBA palette from the specified binary file.
// It returns the parsed Palette and any error encountered while opening or parsing the file.
func LoadBinaryRGBAPalette(filename string) (Palette, error) {
	// TODO: implementation is missing
	p := Palette{}
	return p, nil
}

// LoadTextRGBPalette method loads RGB palette from a text file compatible with
// LoadTextRGBPalette loads a Fractint-style text RGB palette from the named file.
// Each non-empty line is expected to contain three integers (red green blue); values
// are converted to bytes (truncated to 0–255) and appended in that order. Lines that
// do not contain exactly three integers are ignored. It returns the parsed Palette
// and any error encountered while reading the file.
func LoadTextRGBPalette(filename string) (Palette, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var palette Palette

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var red, green, blue int
		items, err := fmt.Sscanf(line, "%d %d %d", &red, &green, &blue)
		if err != nil {
			log.Fatal(err)
		}
		if items != 3 {
			log.Println("not expected line:", line)
		}
		color := []byte{byte(red), byte(green), byte(blue)}
		palette = append(palette, color)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return palette, err
}

// LoadTextRGBAPalette method loads RGBA palette from a text file that is
// LoadTextRGBAPalette loads an RGBA palette from a Fractint-semi-compatible text file.
// Currently unimplemented: it returns an empty Palette and a nil error.
func LoadTextRGBAPalette(filename string) (Palette, error) {
	p := Palette{}
	return p, nil
}