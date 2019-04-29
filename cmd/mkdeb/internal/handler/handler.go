package handler

import (
	"strings"

	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

// Func is an upstream source handler function.
type Func func(*deb.Package, *recipe.Recipe, string, string) error

func stripName(name string, n int) string {
	if n == 0 {
		return name
	}

	count := n
	for count > 0 {
		parts := strings.SplitN(name, "/", 2)
		if len(parts) == 2 {
			name = parts[1]
		}

		count--
	}

	return name
}
