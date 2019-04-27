package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	filetype "gopkg.in/h2non/filetype.v1"
	"mkdeb.sh/cmd/mkdeb/internal/handler"
	"mkdeb.sh/cmd/mkdeb/internal/print"
	"mkdeb.sh/cmd/mkdeb/internal/progress"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
	"mkdeb.sh/repository"
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
		if _, err := exec.LookPath("apt-get"); err != nil {
			return errors.New(`flag "--install" can only be used on Debian-based systems`)
		}
	}

	if ctx.String("to") != "" && ctx.NArg() > 1 {
		return errors.New(`flag "--to" cannot be used when multiple packages are being built`)
	}

	repo := repository.NewRepository(repositoryDir)
	if !repo.Exists() {
		if err := ctx.App.Run([]string{ctx.App.Name, "update"}); err != nil {
			return errors.Wrap(err, "failed to initialize repository")
		}
	}

	for _, arg := range ctx.Args() {
		var (
			recipe *recipe.Recipe
			err    error
		)

		name, arch, version := parseRef(arg)
		if arch == "" {
			arch = "all"
		}
		if version == "" {
			return errors.New("missing recipe version")
		}

		print.Start("Package %s", ansi.Color(name, "green+b"))

		repository := repository.NewRepository(repositoryDir)

		if v := ctx.String("recipe"); v != "" {
			recipe, err = repository.RecipeFromPath(v)
		} else {
			recipe, err = repository.Recipe(name)
		}
		if err != nil {
			return errors.Wrap(err, "cannot load recipe")
		}

		from := ctx.String("from")
		if from == "" {
			from, err = downloadArchive(arch, version, recipe, ctx.Bool("skip-cache"))
			if err != nil {
				return errors.Wrap(err, "cannot download upstream archive")
			}
		}

		print.Step("Using %q upstream archive...", from)

		// Get for output file path (will be overwritten if left empty)
		epoch := ctx.Uint("epoch")
		if epoch == 0 {
			epoch = recipe.Control.Version.Epoch
		}

		to := ctx.String("to")

		info, err := createPackage(arch, version, epoch, ctx.Int("revision"), recipe, from, to)
		if err != nil {
			return errors.Wrap(err, "cannot create package")
		}

		print.Summary("📦", info.String())

		if install {
			pkgs = append(pkgs, info)
		}
	}

	if install && len(pkgs) > 0 {
		print.Start("Install packages")

		if err := installPackages(pkgs); err != nil {
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

func downloadArchive(arch, version string, recipe *recipe.Recipe, force bool) (string, error) {
	var path string

	v, ok := recipe.Source.ArchMapping[arch]
	if !ok {
		return "", errors.New("unsupported architecture")
	}
	arch = v

	// Generate URL from recipe template
	buf := bytes.NewBuffer(nil)

	tmpl, err := template.New("url").Parse(recipe.Source.URL)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse URL template")
	} else if err = tmpl.Execute(buf, struct{ Arch, Version string }{arch, version}); err != nil {
		return "", errors.Wrap(err, "cannot execute URL template")
	}

	url := buf.String()
	idx := strings.LastIndex(url, "/")
	if idx != -1 {
		path = getCachePath(recipe.Name, url[idx+1:])
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	dirPath := filepath.Dir(path)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dirPath, 0755); err != nil {
			return "", errors.Wrap(err, "cannot create cache directory")
		}
	}

	print.Step("Downloading %q...", url)

	req, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	if req.StatusCode >= 400 {
		return "", errors.New(req.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentLength := uint64(req.ContentLength)
	printLength := 0
	progressFunc := func(s uint64) {
		var str string

		if req.ContentLength == -1 || s == contentLength {
			str = fmt.Sprintf("\rdownload %s", humanize.Bytes(s))
		} else {
			str = fmt.Sprintf("\rdownload %s/%s", humanize.Bytes(s), humanize.Bytes(contentLength))
		}

		fmt.Print(str)
		if diff := printLength - len(str); diff > 0 {
			fmt.Print(strings.Repeat(" ", diff))
		}
		printLength = len(str)
	}

	_, err = io.Copy(f, progress.NewReader(req.Body, progressFunc))
	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	return path, nil
}

func createPackage(arch, version string, epoch uint, revision int, recipe *recipe.Recipe, from,
	to string) (*packageInfo, error) {

	var (
		f       handler.Func
		subtype string
	)

	if _, ok := recipe.Source.ArchMapping[arch]; !ok {
		return nil, errUnsupportedArch
	}

	p, err := deb.NewPackage(recipe.Name, arch, version, epoch, revision)
	if err != nil {
		return nil, err
	}

	desc := recipe.Description
	if recipe.Control.Description != "" {
		desc += "\n" + recipe.Control.Description
	}
	p.Control.Set("Description", desc)

	if len(recipe.Control.Depends) > 0 {
		p.Control.Set("Depends", recipe.Control.Depends)
	}
	if len(recipe.Control.PreDepends) > 0 {
		p.Control.Set("PreDepends", recipe.Control.PreDepends)
	}
	if len(recipe.Control.Recommends) > 0 {
		p.Control.Set("Recommends", recipe.Control.Recommends)
	}
	if len(recipe.Control.Suggests) > 0 {
		p.Control.Set("Suggests", recipe.Control.Suggests)
	}
	if len(recipe.Control.Enhances) > 0 {
		p.Control.Set("Enhances", recipe.Control.Enhances)
	}
	if len(recipe.Control.Breaks) > 0 {
		p.Control.Set("Breaks", recipe.Control.Breaks)
	}
	if len(recipe.Control.Conflicts) > 0 {
		p.Control.Set("Conflicts", recipe.Control.Conflicts)
	}

	if len(recipe.Maintainer) > 0 {
		p.Control.Set("Maintainer", recipe.Maintainer)
	}

	if len(recipe.ControlFiles) > 0 {
		print.Step("Adding control files...")

		for _, f := range recipe.ControlFiles {
			name := f.FileInfo.Name()

			fmt.Printf("append %q (%s)\n", name, humanize.Bytes(uint64(f.FileInfo.Size())))

			src, err := os.Open(f.Path)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot open %q file", name)
			}

			if err = p.AddControlFile(name, src, f.FileInfo); err != nil {
				return nil, errors.Wrapf(err, "cannot add %q file", name)
			}
		}
	}

	print.Step("Adding files upstream archive...")

	switch recipe.Source.Type {
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
		return nil, errors.New("unsupported source")
	}

	err = f(p, recipe, from, subtype)
	if err != nil {
		return nil, err
	}

	if len(recipe.RecipeFiles) > 0 {
		print.Step("Adding recipe files...")

		for _, f := range recipe.RecipeFiles {
			name := f.FileInfo.Name()

			if path, confFile, ok := recipe.InstallPath(name, recipe.Install.Recipe); ok {
				fmt.Printf("append %q as %q (%s)\n", name, path, humanize.Bytes(uint64(f.FileInfo.Size())))

				if confFile {
					p.RegisterConfFile(path)
				}

				src, err := os.Open(f.Path)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot open %q file", name)
				}

				if err = p.AddFile(path, src, f.FileInfo); err != nil {
					return nil, errors.Wrapf(err, "cannot add %q file", name)
				}
			}
		}
	}

	if len(recipe.Dirs) > 0 {
		print.Step("Adding recipe directories...")

		for _, path := range recipe.Dirs {
			fmt.Printf("append %q\n", path)

			if err = p.AddDir(path, 0755); err != nil {
				return nil, errors.Wrapf(err, "cannot add %q directory", path)
			}
		}
	}

	if len(recipe.Links) > 0 {
		print.Step("Adding recipe symbolic links...")

		for dst, src := range recipe.Links {
			fmt.Printf("link %q to %q\n", src, dst)

			if err = p.AddLink(dst, src); err != nil {
				return nil, errors.Wrapf(err, "cannot add %q link", dst)
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
			return nil, errors.Wrap(err, "cannot get current directory")
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

	if err := p.Write(file); err != nil {
		return nil, errors.Wrap(err, "cannot write package")
	}

	fi, err := os.Stat(info.Path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get file size")
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

func getCachePath(pkgName, name string) string {
	return filepath.Join(cacheDir, string(pkgName[0]), pkgName, name)
}

type packageInfo struct {
	Path string
	Size int64
}

func (p *packageInfo) String() string {
	return fmt.Sprintf("%s: size=%s", p.Path, humanize.Bytes(uint64(p.Size)))
}
