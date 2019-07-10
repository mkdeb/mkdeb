package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	humanize "github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/mgutz/ansi"
	"github.com/urfave/cli"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"
	"mkdeb.sh/catalog"
	"mkdeb.sh/cmd/mkdeb/internal/handler"
	"mkdeb.sh/cmd/mkdeb/internal/print"
	"mkdeb.sh/cmd/mkdeb/internal/progress"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

var buildCommand = cli.Command{
	Name:      "build",
	Usage:     "Build Debian package",
	Action:    execBuild,
	ArgsUsage: "RECIPE...",
	Flags: []cli.Flag{
		cli.UintFlag{
			Name:  "epoch, e",
			Usage: "Package version epoch",
		},
		cli.StringFlag{
			Name:  "from, f",
			Usage: "Upstream archive path",
		},
		cli.BoolFlag{
			Name:  "install, i",
			Usage: "Install package after build",
		},
		cli.StringFlag{
			Name:  "recipe, R",
			Usage: "Recipe base path",
		},
		cli.IntFlag{
			Name:  "revision, r",
			Usage: "Package version revision",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "skip-cache",
			Usage: "Skip download cache",
		},
		cli.StringFlag{
			Name:  "to, t",
			Usage: "Package output path",
		},
	},
}

func execBuild(ctx *cli.Context) error {
	var pkgs []*packageInfo

	if ctx.NArg() == 0 {
		cli.ShowCommandHelpAndExit(ctx, "build", 1)
	}

	install := ctx.Bool("install")
	if install {
		_, err := exec.LookPath("apt-get")
		if err != nil {
			return xerrors.New(`flag "--install" can only be used on Debian-based systems`)
		}
	}

	if ctx.String("to") != "" && ctx.NArg() > 1 {
		return xerrors.New(`flag "--to" cannot be used when multiple packages are being built`)
	}

	if !catalog.Ready(catalogDir) {
		err := ctx.App.Run([]string{ctx.App.Name, "repo", "add", catalog.DefaultRepository})
		if err != nil {
			return err
		}
	}

	c, err := catalog.New(catalogDir)
	if err != nil {
		return xerrors.Errorf("cannot initialize catalog: %w", err)
	}
	defer c.Close()

	for _, arg := range ctx.Args() {
		var (
			rcp *recipe.Recipe
			err error
		)

		name, arch, version := parseRef(arg)
		if arch == "" {
			arch = "all"
		}
		if version == "" {
			return xerrors.New("missing recipe version")
		}

		rcpPath := ctx.String("recipe")
		if rcpPath != "" {
			rcp, err = recipe.LoadRecipe(rcpPath)
		} else {
			rcp, err = c.Recipe(name)
		}
		if err == catalog.ErrRecipeNotFound {
			return err
		} else if err != nil {
			return xerrors.Errorf("cannot load recipe: %w", err)
		}

		print.Section("Package %s", ansi.Color(name, "green+b"))

		from := ctx.String("from")
		if from == "" {
			from, err = downloadArchive(arch, version, rcp, ctx.Bool("skip-cache"))
			if err != nil {
				return xerrors.Errorf("cannot download upstream archive: %w", err)
			}
		} else {
			rcp.Source.URL = "<unused>"
			rcp.Source.Type = "file"
		}

		fi, err := os.Stat(from)
		if err == nil && fi.IsDir() {
			print.Step("Using %q upstream folder...", from)
		} else {
			print.Step("Using %q upstream file...", from)
		}

		// Get for output file path (will be overwritten if left empty)
		epoch := ctx.Uint("epoch")
		if epoch == 0 {
			epoch = rcp.Control.Version.Epoch
		}

		to := ctx.String("to")

		// Ensure recipe is valid before build
		err = rcp.Validate()
		if err != nil {
			return err
		}

		info, err := createPackage(arch, version, epoch, ctx.Int("revision"), rcp, from, to)
		if err != nil {
			return xerrors.Errorf("cannot create package: %w", err)
		}

		print.Summary("📦", info.String())

		if install {
			pkgs = append(pkgs, info)
		}
	}

	if install && len(pkgs) > 0 {
		print.Section("Install packages")

		err := installPackages(pkgs)
		if err != nil {
			return err
		}

		message.Set(language.English, "build.install", plural.Selectf(1, "%d",
			plural.One, "Operation installed %d package",
			plural.Other, "Operation installed %d packages",
		))

		print.Summary("📋", message.NewPrinter(language.English).Sprintf("build.install", len(pkgs)))
	}

	return nil
}

func parseRef(input string) (string, string, string) {
	var name, arch, version string

	if strings.Contains(input, "=") {
		parts := strings.SplitN(input, "=", 2)
		input = parts[0]
		version = parts[1]
	}

	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		name = parts[0]
		arch = parts[1]
	} else {
		name = input
	}

	return name, arch, version
}

func downloadArchive(arch, version string, rcp *recipe.Recipe, force bool) (string, error) {
	var path string

	v, ok := rcp.Source.ArchMapping[arch]
	if !ok {
		return "", xerrors.New("unsupported architecture")
	}
	arch = v

	// Generate URL from recipe template
	buf := bytes.NewBuffer(nil)

	tmpl, err := template.New("").Parse(rcp.Source.URL)
	if err != nil {
		return "", xerrors.Errorf("cannot parse URL template: %w", err)
	} else if err = tmpl.Execute(buf, struct{ Arch, Version string }{arch, version}); err != nil {
		return "", xerrors.Errorf("cannot execute URL template: %w", err)
	}

	url := buf.String()
	idx := strings.LastIndex(url, "/")
	if idx != -1 {
		path = filepath.Join(cacheDir, string(rcp.Name[0]), rcp.Name, url[idx+1:])
	}

	if !force {
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
	}

	dirPath := filepath.Dir(path)
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dirPath, 0755); err != nil {
			return "", xerrors.Errorf("cannot create cache directory: %w", err)
		}
	}

	print.Step("Downloading %q...", url)

	req, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	if req.StatusCode >= 400 {
		return "", xerrors.New(req.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentLength := uint64(req.ContentLength)
	printLength := 0
	progressFn := func(s uint64) {
		var str string

		if req.ContentLength == -1 || s == contentLength {
			str = fmt.Sprintf("\rdownload %s", humanize.Bytes(s))
		} else {
			str = fmt.Sprintf("\rdownload %s/%s", humanize.Bytes(s), humanize.Bytes(contentLength))
		}

		fmt.Print(str)
		diff := printLength - len(str)
		if diff > 0 {
			fmt.Print(strings.Repeat(" ", diff))
		}
		printLength = len(str)
	}

	_, err = io.Copy(f, progress.New(req.Body, progressFn))
	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	return path, nil
}

func createPackage(arch, version string, epoch uint, revision int, rcp *recipe.Recipe, from,
	to string) (*packageInfo, error) {

	var (
		f       handler.Func
		subtype string
	)

	_, ok := rcp.Source.ArchMapping[arch]
	if !ok {
		return nil, xerrors.New("unsupported architecture")
	}

	p, err := deb.NewPackage(rcp.Name, arch, version, epoch, revision)
	if err != nil {
		return nil, err
	}

	desc := rcp.Description
	if rcp.Control.Description != "" {
		desc += "\n" + rcp.Control.Description
	}
	p.Control.Description = desc

	if len(rcp.Control.Depends) > 0 {
		p.Control.Depends = rcp.Control.Depends
	}
	if len(rcp.Control.PreDepends) > 0 {
		p.Control.PreDepends = rcp.Control.PreDepends
	}
	if len(rcp.Control.Recommends) > 0 {
		p.Control.Recommends = rcp.Control.Recommends
	}
	if len(rcp.Control.Suggests) > 0 {
		p.Control.Suggests = rcp.Control.Suggests
	}
	if len(rcp.Control.Enhances) > 0 {
		p.Control.Enhances = rcp.Control.Enhances
	}
	if len(rcp.Control.Breaks) > 0 {
		p.Control.Breaks = rcp.Control.Breaks
	}
	if len(rcp.Control.Conflicts) > 0 {
		p.Control.Conflicts = rcp.Control.Conflicts
	}

	if len(rcp.Maintainer) > 0 {
		p.Control.Maintainer = rcp.Maintainer
	}

	if len(rcp.ControlFiles) > 0 {
		print.Step("Adding control files...")

		for _, f := range rcp.ControlFiles {
			name := f.FileInfo.Name()

			fmt.Printf("append %q (%s)\n", name, humanize.Bytes(uint64(f.FileInfo.Size())))

			src, err := os.Open(f.Path)
			if err != nil {
				return nil, xerrors.Errorf("cannot open %q file: %w", name, err)
			}

			if err = p.AddControlFile(name, src, f.FileInfo); err != nil {
				return nil, xerrors.Errorf("cannot add %q file: %w", name, err)
			}
		}
	}

	print.Step("Adding upstream files...")

	switch rcp.Source.Type {
	case "archive":
		typ, err := filetype.MatchFile(from)
		if err != nil {
			return nil, err
		}

		switch typ.MIME.Subtype {
		case "gzip", "x-bzip2", "x-tar", "x-xz":
			f = handler.Tar

		case "zip":
			f = handler.Zip
		}

		subtype = typ.MIME.Subtype

	case "file":
		f = handler.File
	}

	if f == nil {
		return nil, xerrors.New("unsupported source")
	}

	err = f(p, rcp, from, subtype)
	if err != nil {
		return nil, err
	}

	if len(rcp.RecipeFiles) > 0 {
		print.Step("Adding recipe files...")

		for _, f := range rcp.RecipeFiles {
			name := f.FileInfo.Name()

			path, confFile, ok := rcp.InstallPath(name, rcp.Install.Recipe)
			if ok {
				fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(f.FileInfo.Size())))

				if confFile {
					p.RegisterConfFile(path)
				}

				src, err := os.Open(f.Path)
				if err != nil {
					return nil, xerrors.Errorf("cannot open %q file: %w", name, err)
				}

				if err = p.AddFile(path, src, f.FileInfo); err != nil {
					return nil, xerrors.Errorf("cannot add %q file: %w", name, err)
				}
			}
		}
	}

	if len(rcp.Dirs) > 0 {
		print.Step("Adding recipe directories...")

		for _, path := range rcp.Dirs {
			fmt.Printf("append %q\n", path)

			if err = p.AddDir(path, 0755); err != nil {
				return nil, xerrors.Errorf("cannot add %q directory: %w", path, err)
			}
		}
	}

	if len(rcp.Links) > 0 {
		print.Step("Adding recipe symbolic links...")

		for dst, src := range rcp.Links {
			fmt.Printf("link %q to %q\n", src, dst)

			if err = p.AddLink(dst, src); err != nil {
				return nil, xerrors.Errorf("cannot add %q link: %w", dst, err)
			}
		}
	}

	// Set default file path is empty
	if to == "" {
		v := p.Version.Upstream
		if p.Version.Revision != "" {
			v += "-" + p.Version.Revision
		}

		wd, err := os.Getwd()
		if err != nil {
			return nil, xerrors.Errorf("cannot get current directory: %w", err)
		}

		to = filepath.Join(wd, fmt.Sprintf("%s_%s_%s.deb", p.Name, v, p.Arch))
	}

	info := &packageInfo{
		Path: to,
	}

	file, err := os.Create(info.Path)
	if err != nil {
		return nil, err
	}

	err = p.Write(file)
	if err != nil {
		return nil, xerrors.Errorf("cannot write package: %w", err)
	}

	fi, err := os.Stat(info.Path)
	if err != nil {
		return nil, xerrors.Errorf("cannot get file size: %w", err)
	}

	info.Size = fi.Size()

	return info, nil
}

func installPackages(pkgs []*packageInfo) error {
	var paths []string

	for _, pkg := range pkgs {
		paths = append(paths, pkg.Path)
	}

	apt, err := exec.LookPath("apt-get")
	if err != nil {
		return err
	}

	cmd := exec.Command(apt, append([]string{"install", "--reinstall", "-y"}, paths...)...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

type packageInfo struct {
	Path string
	Size int64
}

func (p *packageInfo) String() string {
	return fmt.Sprintf("%s: size=%s", p.Path, humanize.Bytes(uint64(p.Size)))
}
