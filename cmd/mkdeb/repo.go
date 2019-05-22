package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/mgutz/ansi"
	"github.com/urfave/cli"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"
	"mkdeb.sh/catalog"
	"mkdeb.sh/cmd/mkdeb/internal/print"
)

var repoCommand = cli.Command{
	Name:  "repo",
	Usage: "Manage recipes repositories",
	Subcommands: []cli.Command{
		{
			Name:      "add",
			Usage:     "Install new repository",
			ArgsUsage: "user/repository [URL]",
			Action:    execAdd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "branch",
					Usage: "Repository branch name",
					Value: "master",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "Force repository installation",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List installed repositories",
			Action: execList,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format",
					Usage: "Output template format",
				},
			},
		},
		{
			Name:   "remove",
			Usage:  "Remove installed repository",
			Action: execRemove,
		},
	},
}

func execAdd(ctx *cli.Context) error {
	if ctx.NArg() < 1 || ctx.NArg() > 2 {
		cli.ShowCommandHelpAndExit(ctx, "add", 1)
	}

	name := ctx.Args().Get(0)
	if strings.Count(name, "/") != 1 {
		return fmt.Errorf(`invalid %q repository name, must match "user/repository"`, name)
	}

	branch := ctx.String("branch")

	url := ctx.Args().Get(1)
	if url == "" {
		parts := strings.SplitN(name, "/", 2)
		url = fmt.Sprintf("https://github.com/%s/mkdeb-%s.git", parts[0], parts[1])
	} else if strings.Contains(url, "@") {
		parts := strings.SplitN(url, "@", 2)
		url, branch = parts[0], parts[1]
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return xerrors.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	print.Section("Repository %s", ansi.Color(name, "green+b"))

	print.Step("Installing %s repository...", url)

	count, err := c.InstallRepository(name, url, branch, ctx.Bool("force"))
	if err == catalog.ErrRepositoryExist {
		return xerrors.New(`repository already installed, use "--force" to reinstall`)
	} else if err != nil {
		return xerrors.Errorf("cannot install repository: %w", err)
	}

	message.Set(language.English, "update.result", plural.Selectf(1, "%d",
		plural.One, "%d recipe",
		plural.Other, "%d recipes",
	))

	print.Summary("📋", "Repository installed: "+message.NewPrinter(language.English).Sprintf("update.result", count))

	return nil
}

func execList(ctx *cli.Context) error {
	c, err := catalog.New(catalogDir)
	if err != nil {
		return xerrors.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	repos, err := c.Repositories()
	if err != nil {
		return err
	}

	format := ctx.String("format")
	if format == "" {
		format = "{{ .Name }}\t{{ .URL }}@{{ .Branch }}\n"
	} else {
		format = strings.TrimSpace(format) + "\n"
	}

	tmpl, err := template.New("").Parse(format)
	if err != nil {
		return xerrors.Errorf("invalid format: %w", err)
	}

	tr := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, repo := range repos {
		err = tmpl.Execute(tr, repo)
		if err != nil {
			return xerrors.Errorf("cannot execute template: %w", err)
		}
	}
	tr.Flush()

	return nil
}

func execRemove(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		cli.ShowCommandHelpAndExit(ctx, "remove", 1)
	}

	name := ctx.Args().Get(0)
	if strings.Count(name, "/") != 1 {
		return fmt.Errorf(`invalid %q repository name, must match "user/repository"`, name)
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return xerrors.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	print.Section("Repository %s", ansi.Color(name, "green+b"))
	print.Step("Uninstalling repository...")

	count, err := c.UninstallRepository(name)
	if err != nil {
		return xerrors.Errorf("cannot uninstall repository: %w", err)
	}

	message.Set(language.English, "update.result", plural.Selectf(1, "%d",
		plural.One, "%d recipe",
		plural.Other, "%d recipes",
	))

	print.Summary("🗑", "Repository uninstalled: "+message.NewPrinter(language.English).
		Sprintf("update.result", count))

	return nil
}
