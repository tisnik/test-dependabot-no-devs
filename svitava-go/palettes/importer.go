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

// LoadBinaryRGBPalette method loads RGB palette from binary file.
func LoadBinaryRGBPalette(filename string) (Palette, error) {
	// TODO: implementation is missing
	p := Palette{}
	return p, nil
}

// LoadBinaryRGBPalette method loads RGBA palette from binary file.
func LoadBinaryRGBAPalette(filename string) (Palette, error) {
	// TODO: implementation is missing
	p := Palette{}
	return p, nil
}

// LoadTextRGBPalette method loads RGB palette from a text file compatible with
// Fractint.
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
// semi-compatible with Fractint.
func LoadTextRGBAPalette(filename string) (Palette, error) {
	p := Palette{}
	return p, nil
}
