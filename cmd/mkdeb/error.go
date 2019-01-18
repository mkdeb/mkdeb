package main

import "errors"

var (
	errEmptyVersion      = errors.New("empty version")
	errUnsupportedArch   = errors.New("unsupported architecture")
	errUnsupportedSource = errors.New("unsupported source")
)
