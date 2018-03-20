package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli"
)

var (
	configDir     string
	cacheDir      string
	repositoryDir string
)

func main() {
	if err := run(); err != nil {
		printError("%s", err)
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
		updateCommand,
		versionCommand,
	}

	return app.Run(os.Args)
}
