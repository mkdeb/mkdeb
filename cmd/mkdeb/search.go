package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"facette.io/natsort"
	_ "github.com/blevesearch/bleve/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"mkdeb.sh/catalog"
)

var searchCommand = cli.Command{
	Name:      "search",
	Usage:     "Search for recipes",
	Action:    execSearch,
	ArgsUsage: "[TERM]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "description, desc",
			Usage: "Include recipes description when searching",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "Output template format",
		},
	},
}

func execSearch(ctx *cli.Context) error {
	if ctx.NArg() > 1 {
		cli.ShowCommandHelpAndExit(ctx, "search", 1)
	}

	if !catalog.Ready(catalogDir) {
		err := ctx.App.Run([]string{ctx.App.Name, "repo", "add", catalog.DefaultRepository})
		if err != nil {
			return err
		}
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return errors.Wrap(err, "cannot initialize catalog")
	}
	defer c.Close()

	term := ctx.Args().Get(0)

	hits, err := c.Search(term, ctx.Bool("desc"))
	if err != nil {
		return errors.Wrap(err, "cannot search catalog")
	}

	if len(hits) == 0 {
		if term != "" {
			fmt.Printf("No match found for %q\n", term)
		} else {
			fmt.Println("No match found")
		}
		return nil
	}

	sort.Slice(hits, func(i, j int) bool {
		return natsort.Compare(hits[i].Repository+"/"+hits[i].Name, hits[j].Repository+"/"+hits[j].Name)
	})

	format := "{{ if ne .Repository \"" + catalog.DefaultRepository +
		"\" }}{{ .Repository }}/{{ end }}{{ .Name }}\t{{ .Description }}\n"
	if v := ctx.String("format"); v != "" {
		format = strings.TrimSpace(v) + "\n"
	}

	tmpl, err := template.New("").Parse(format)
	if err != nil {
		return errors.Wrap(err, "invalid format")
	}

	tr := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, hit := range hits {
		err = tmpl.Execute(tr, hit)
		if err != nil {
			return errors.Wrap(err, "cannot execute template")
		}
	}
	tr.Flush()

	return nil
}
