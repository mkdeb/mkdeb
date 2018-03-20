package main

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli"
)

var version = "0.1.0dev"

var versionCommand = cli.Command{
	Name:   "version",
	Usage:  "print version information",
	Action: execVersion,
}

func execVersion(ctx *cli.Context) error {
	fmt.Printf(
		"%s v%s %s/%s %s/%s\n",
		ctx.App.Name,
		version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		runtime.Compiler,
	)

	return nil
}