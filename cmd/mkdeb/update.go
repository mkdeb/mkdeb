package main

import (
	"fmt"
	"os"

	"mkdeb.sh/cmd/mkdeb/internal/print"

	"github.com/urfave/cli"
	"mkdeb.sh/repository"
)

var updateCommand = cli.Command{
	Name:   "update",
	Usage:  "update recipes repository",
	Action: execUpdate,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "reset",
			Usage: "force repository reset",
		},
	},
}

func execUpdate(ctx *cli.Context) error {
	r := repository.NewRepository(repositoryDir)
	if !r.Exists() {
		print.Step("Initializing repository...")
		return r.Init(os.Stdout)
	}

	print.Step("Updating repository...")

	err := r.Update(os.Stdout, ctx.Bool("reset"))
	if err == repository.ErrAlreadyUpToDate {
		fmt.Println(err)
		return nil
	}

	return err
}
