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
	"net/http"
	"net/url"
	"strconv"
)

// parseStringQueryParameter retrieves the first value for the named query parameter from r.URL.RawQuery or returns defaultValue if the parameter is not present.
// If the raw query cannot be parsed, it returns defaultValue and the parse error.
func parseStringQueryParameter(
	r *http.Request,
	parameter string,
	defaultValue string) (string, error) {

	urlParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return defaultValue, err
	}

	if paramStr, found := urlParams[parameter]; found {
		return paramStr[0], nil
	}

	return defaultValue, nil
}

// parseUintQueryParameter retrieves an unsigned integer from the request URL query.
// It reads the value of parameterName from r.URL.Query(); if the parameter is absent
// it returns defaultValue and no error. If the parameter is present but cannot be
// parsed as a base-10 unsigned integer, it returns defaultValue and the parse error.
func parseUintQueryParameter(
	r *http.Request,
	parameterName string,
	defaultValue uint) (uint, error) {

	valueAsString := r.URL.Query().Get(parameterName)

	// parameter not present
	if valueAsString == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseUint(valueAsString, 10, 0)
	if err != nil {
		return defaultValue, err
	}
	return uint(value), nil
}