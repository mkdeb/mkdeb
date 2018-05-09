package main

import (
	"os"
	"os/user"
	"path/filepath"

	"mkdeb.sh/cmd/mkdeb/internal/print"

	"github.com/urfave/cli"
)

var (
	configDir     string
	cacheDir      string
	indexDir      string
	repositoryDir string
)

func main() {
	if err := run(); err != nil {
		print.Error("%s", err)
		os.Exit(1)
	}
}

func run() error {
	// Get directories paths
	usr, err := user.Current()
	if err != nil {
		return err
	}

	configDir = filepath.Join(usr.HomeDir, ".mkdeb")
	cacheDir = filepath.Join(configDir, "cache")
	indexDir = filepath.Join(configDir, "recipes/index")
	repositoryDir = filepath.Join(configDir, "recipes/default")

	// Create cache directory if missing
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}

	// Initialize application CLI
	app := cli.NewApp()
	app.Name = "mkdeb"
	app.Usage = "The Debian packaging helper"
	app.HideVersion = true

	app.Commands = []cli.Command{
		buildCommand,
		cleanCommand,
		searchCommand,
		updateCommand,
		versionCommand,
	}

	return app.Run(os.Args)
}
