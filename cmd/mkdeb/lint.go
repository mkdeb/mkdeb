package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"mkdeb.sh/catalog"
	"mkdeb.sh/cmd/mkdeb/internal/print"
	"mkdeb.sh/lint"
	"mkdeb.sh/recipe"
)

var lintCommand = cli.Command{
	Name:      "lint",
	Usage:     "Run linter on recipes",
	Action:    execLint,
	ArgsUsage: "(--tags TAGS...|[RECIPE...])",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "tags",
			Usage: "Display linting rule tags information",
		},
	},
}

func execLint(ctx *cli.Context) error {
	var failed bool

	if ctx.Bool("tags") {
		for _, arg := range ctx.Args() {
			print.LintInfo(lint.Info(arg))
		}

		return nil
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return errors.Wrap(err, "cannot initialize catalog")
	}
	defer c.Close()

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args() {
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
		c.Walk(func(rcp *recipe.Recipe, repo *catalog.Repository, err error) error {
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
	}

	if failed {
		os.Exit(1)
	}

	return nil
}
