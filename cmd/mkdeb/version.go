package main

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

var version = "0.4.0dev"

var versionCommand = &cli.Command{
	Name:      "version",
	Usage:     "Print version information",
	ArgsUsage: " ",
	Action:    execVersion,
}

func execVersion(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		cli.ShowCommandHelpAndExit(ctx, "version", 1)
	}

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
