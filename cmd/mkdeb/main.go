package main

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var (
	dataDir    string
	cacheDir   string
	catalogDir string
)

func main() {
	cli.HelpFlag = helpFlag
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate

	app := &cli.App{
		Name:  "mkdeb",
		Usage: "Debian packaging helper",
		Commands: []*cli.Command{
			buildCommand,
			cleanupCommand,
			helpCommand,
			lintCommand,
			repoCommand,
			searchCommand,
			updateCommand,
			versionCommand,
		},
		Flags: []cli.Flag{
			helpFlag,
			&cli.BoolFlag{
				Name:  "no-emoji",
				Usage: "Disable emoji in commands output",
			},
		},
		HideHelp:    true,
		HideVersion: true,
		Before: func(ctx *cli.Context) error {
			// Set base directories paths
			dir, err := homedir.Dir()
			if err != nil {
				return err
			}

			dataDir = filepath.Join(dir, ".mkdeb")
			cacheDir = filepath.Join(dataDir, "cache")
			catalogDir = filepath.Join(dataDir, "catalog")

			if ctx.Bool("no-emoji") {
				print.DisableEmoji()
			}

			return nil
		},
		Action: cli.ShowAppHelp,
	}

	err := app.Run(os.Args)
	if err != nil {
		print.Error("%s", err)
		os.Exit(1)
	}
}
