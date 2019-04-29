package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"mkdeb.sh/catalog"
	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var updateCommand = cli.Command{
	Name:      "update",
	Usage:     "Update recipes repositories",
	Action:    execUpdate,
	ArgsUsage: " ",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force",
			Usage: "Force actions and repair any dangling state",
		},
		cli.BoolFlag{
			Name:  "index-only",
			Usage: "Only perform recipes indexing",
		},
	},
}

func execUpdate(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		cli.ShowCommandHelpAndExit(ctx, "update", 1)
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return errors.Wrap(err, "cannot initialize catalog")
	}
	defer c.Close()

	repos, err := c.Repositories()
	if err != nil {
		return errors.Wrap(err, "cannot get repositories")
	}

	if len(repos) == 0 {
		fmt.Println("No repository installed")
		return nil
	}

	if !ctx.Bool("index-only") {
		for _, repo := range repos {
			print.Step("Updating %s repository...", repo.Name)

			err = repo.Update(ctx.Bool("force"))
			if err == catalog.ErrAlreadyUpToDate {
				fmt.Println(err)
			} else if err != nil {
				return errors.Wrap(err, "cannot update repository")
			}
		}
	}

	count, err := c.Index()
	if err != nil {
		return errors.Wrap(err, "cannot index repositories")
	}

	message.Set(language.English, "update.result", plural.Selectf(1, "%d",
		plural.One, "indexed %d recipe",
		plural.Other, "indexed %d recipes",
	))

	print.Summary("📋", message.NewPrinter(language.English).Sprintf("update.result", count))

	return nil
}
