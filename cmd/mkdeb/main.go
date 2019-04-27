package main

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var (
	dataDir       string
	cacheDir      string
	indexDir      string
	repositoryDir string
)

func main() {
	err := run()
	if err != nil {
		print.Error("%s", err)
		os.Exit(1)
	}
}

func run() error {
	// Set base directories paths
	dir, err := homedir.Dir()
	if err != nil {
		return err
	}

	dataDir = filepath.Join(dir, ".mkdeb")
	cacheDir = filepath.Join(dataDir, "cache")
	indexDir = filepath.Join(dataDir, "recipes/index")
	repositoryDir = filepath.Join(dataDir, "recipes/default")

	// Create cache directory if missing
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}

	// Run CLI application
	cli.HelpFlag = helpFlag
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate

	app := cli.NewApp()
	app.Name = "mkdeb"
	app.Usage = "Debian packaging helper"
	app.Commands = []cli.Command{
		buildCommand,
		cleanCommand,
		helpCommand,
		searchCommand,
		updateCommand,
		versionCommand,
	}
	app.Flags = []cli.Flag{
		helpFlag,
		cli.BoolFlag{
			Name:  "no-emoji",
			Usage: "Disable emoji in commands output",
		},
	}
	app.HideHelp = true
	app.HideVersion = true
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("no-emoji") {
			print.DisableEmoji()
		}

		return nil
	}

	return app.Run(os.Args)
}
