package main

import (
	"fmt"
	"os"
	"path/filepath"

	humanize "github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"

	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var cleanupCommand = &cli.Command{
	Name:      "cleanup",
	Usage:     "Cleanup local cache",
	ArgsUsage: " ",
	Action:    execCleanup,
}

func execCleanup(ctx *cli.Context) error {
	var size int64

	print.Section("Cleanup")
	print.Step("Removing files from cache...")

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot path %q path: %w", path, err)
		}

		if !info.IsDir() {
			sz := info.Size()

			fmt.Printf("remove %q file (%s)\n", path, humanize.Bytes(uint64(sz)))

			if err = os.Remove(path); err != nil {
				return fmt.Errorf("cannot delete %q file: %w", path, err)
			}

			size += sz
		}

		return nil
	})

	print.Summary("🗑", "Operation freed %s of disk space", humanize.Bytes(uint64(size)))

	return err
}
