package main

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"mkdeb.sh/cmd/mkdeb/internal/print"
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
	dir, err := homedir.Dir()
	if err != nil {
		return err
	}

	configDir = filepath.Join(dir, ".mkdeb")
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
