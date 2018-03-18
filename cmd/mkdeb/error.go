package main

import "errors"

var (
	errEmptyVersion    = errors.New("empty version")
	errRecipeNotFound  = errors.New("recipe not found")
	errUnsupportedArch = errors.New("unsupported architecture")
)
