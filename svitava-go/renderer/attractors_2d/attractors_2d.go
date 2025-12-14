//
//  (C) Copyright 2025  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

// 2D attractors
package attractors_2d

// sign returns 1 if x is greater than zero; otherwise it returns 0.
func sign(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}