package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"mkdeb.sh/catalog"
	"mkdeb.sh/lint"
	"mkdeb.sh/recipe"

	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var lintCommand = &cli.Command{
	Name:      "lint",
	Usage:     "Run linter on recipes",
	ArgsUsage: "(--tags TAGS...|[RECIPE...])",
	Action:    execLint,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "tags",
			Usage: "Display linting rule tags information",
		},
	},
}

func execLint(ctx *cli.Context) error {
	var failed bool

	if ctx.Bool("tags") {
		for _, arg := range ctx.Args().Slice() {
			print.LintInfo(lint.Info(arg))
		}

		return nil
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return fmt.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args().Slice() {
			rcp, err := c.Recipe(arg)
			if err != nil {
				return err
			}

			problems, ok := lint.Lint(rcp)
			if !ok {
				failed = true
			}
			print.Lint(rcp, problems)
		}
	} else {
		err = c.Walk(func(rcp *recipe.Recipe, repo *catalog.Repository, err error) error {
			if err != nil {
				return err
			}

			problems, ok := lint.Lint(rcp)
			if !ok {
				failed = true
			}
			print.Lint(rcp, problems)

			return nil
		})
		if err != nil {
			return fmt.Errorf("cannot walk recipes: %w", err)
		}
	}

	if failed {
		os.Exit(1)
	}

	return nil
}
