//
//  (C) Copyright 2024, 2025  Pavel Tisnovsky
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

// Palette represents color palette used to map fractal calculation result
// (number of iterations, for example) into RGB or RGBA color. Palettes have
// usually 256 records, but it can be more or less.
type Palette [][]byte
