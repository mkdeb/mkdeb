package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"mkdeb.sh/catalog"

	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var updateCommand = &cli.Command{
	Name:      "update",
	Usage:     "Update recipes repositories",
	ArgsUsage: " ",
	Action:    execUpdate,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Force actions and repair any dangling state",
		},
		&cli.BoolFlag{
			Name:  "index-only",
			Usage: "Only perform recipes indexing",
		},
	},
}

func execUpdate(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		cli.ShowCommandHelpAndExit(ctx, "update", 1)
	}

	if !catalog.Ready(catalogDir) {
		err := ctx.App.Run([]string{ctx.App.Name, "repo", "add", catalog.DefaultRepository})
		if err != nil {
			return err
		}
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return fmt.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	repos, err := c.Repositories()
	if err != nil {
		return err
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
				return fmt.Errorf("cannot update repository: %w", err)
			}
		}
	}

	count, err := c.Index()
	if err != nil {
		return fmt.Errorf("cannot index repositories: %w", err)
	}

	message.Set(language.English, "update.result", plural.Selectf(1, "%d",
		plural.One, "indexed %d recipe",
		plural.Other, "indexed %d recipes",
	))

	print.Summary("📋", message.NewPrinter(language.English).Sprintf("update.result", count))

	return nil
}
