package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"./internal/handler"
	"./internal/print"

	humanize "github.com/dustin/go-humanize"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	filetype "gopkg.in/h2non/filetype.v1"
	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
	"mkdeb.sh/repository"
)

var buildCommand = cli.Command{
	Name:   "build",
	Usage:  "build Debian package",
	Action: execBuild,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "from, f",
			Usage: "upstream archive path",
		},
		cli.IntFlag{
			Name:  "revision, r",
			Usage: "package version revision",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "skip-cache",
			Usage: "skip download cache",
		},
		cli.StringFlag{
			Name:  "to, t",
			Usage: "package output path",
		},
	},
}

func execBuild(ctx *cli.Context) error {
	if ctx.String("to") != "" && ctx.NArg() > 1 {
		return errors.New("can't use \"--to\" flag when multiple packages are being built")
	}

	for _, arg := range ctx.Args() {
		name, arch, version := parseRef(arg)
		if arch == "" {
			arch = "all"
		}
		if version == "" {
			return errEmptyVersion
		}

		print.Start("Package %s", ansi.Color(name, "green+b"))

		recipe, err := repository.NewRepository(repositoryDir).Recipe(name)
		if err != nil {
			return errors.Wrap(err, "failed to load recipe")
		}

		from := ctx.String("from")
		if from == "" {
			from, err = downloadArchive(arch, version, recipe, ctx.Bool("skip-cache"))
			if err != nil {
				return errors.Wrap(err, "failed to download upstream archive")
			}
		}

		print.Step("Using %q upstream archive...", from)

		// Get for output file path (will be overwritten if left empty)
		to := ctx.String("to")

		info, err := createPackage(arch, version, ctx.Int("revision"), recipe, from, to)
		if err != nil {
			return errors.Wrap(err, "failed to create package")
		}

		print.Step("Result")
		fmt.Printf("📦   %s\n", info)
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

	_, err = io.Copy(f, req.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}

func createPackage(arch, version string, revision int, recipe *recipe.Recipe, from, to string) (*packageInfo, error) {
	var f handler.Func

	if arch != "" {
		v, ok := recipe.Source.ArchMapping[arch]
		if !ok {
			return nil, errUnsupportedArch
		}
		arch = v
	}

	p, err := deb.NewPackage(recipe.Name, arch, version, revision)
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

	typ, err := filetype.MatchFile(from)
	if err != nil {
		return nil, err
	}

	switch typ.MIME.Subtype {
	case "gzip", "x-bzip2", "x-xz":
		f = handler.Tar

	case "zip":
		f = handler.Zip
	}

	err = f(p, recipe, from, typ.MIME.Subtype)
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
		to = fmt.Sprintf("%s_%s_%s.deb", p.Name, p.Version, p.Arch)
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
