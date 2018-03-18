package main

import (
	"fmt"
	"os"
	"path/filepath"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var cleanCommand = cli.Command{
	Name:   "clean",
	Usage:  "clean packaging cache",
	Action: execClean,
}

func execClean(ctx *cli.Context) error {
	var size int64

	printStart("Clean")

	printStep("Removing files from cache...")

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "cannot path %q path", path)
		}

		if !info.IsDir() {
			sz := info.Size()

			fmt.Printf("remove %q file (%s)\n", path, humanize.Bytes(uint64(sz)))

			if err = os.Remove(path); err != nil {
				return errors.Wrapf(err, "cannot delete %q file", path)
			}

			size += sz
		}

		return nil
	})

	printStep("Result")
	fmt.Printf("🗑   Operation freed %s of disk space\n", humanize.Bytes(uint64(size)))

	return err
}
