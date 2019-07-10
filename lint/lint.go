package lint

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	"mkdeb.sh/deb"
	"mkdeb.sh/recipe"
)

//go:generate go run internal/generate/main.go -o rules.go

// Levels:
const (
	_ = iota
	LevelError
	LevelWarning
)

// Info returns information about a linting rule.
func Info(tag string) *RuleInfo {
	return rules[tag]
}

// Lint inspects recipe and tries to detect problems or guidelines violations.
func Lint(rcp *recipe.Recipe) ([]*Problem, bool) {
	var failed bool

	l := &linter{}
	l.lintVersion(rcp.Version)
	l.lintName(rcp.Name)
	l.lintDescription(rcp.Description)
	l.lintMaintainer(rcp.Maintainer)
	l.lintHomepage(rcp.Homepage)
	l.lintSource(rcp.Source)
	l.lintControl(rcp.Control)
	l.lintInstall(rcp.Install)
	l.lintDirs(rcp.Dirs)
	l.lintLinks(rcp.Links)

	for _, p := range l.problems {
		if p.Level == LevelError {
			failed = true
			break
		}
	}

	return l.problems, !failed
}

// Problem is a linting problem.
type Problem struct {
	Level int
	Tag   string
	Args  []interface{}
}

// RuleInfo is linting rule information.
type RuleInfo struct {
	Tag         string
	Level       int
	Description string
}

type linter struct {
	problems []*Problem
}

func (l *linter) emit(tag string, args ...interface{}) {
	r, ok := rules[tag]
	if !ok {
		panic(fmt.Sprintf("unsupported %q rule", tag))
	}

	l.problems = append(l.problems, &Problem{r.Level, tag, args})
}

func (l *linter) lintVersion(v int) {
	if !recipe.VersionSupported(v) {
		l.emit("version-unsupported", v)
	}
}

func (l *linter) lintName(v string) {
	if v == "" {
		l.emit("name-empty")
		return
	}

	lastIdx := len(v) - 1
	for idx, b := range v {
		if !((b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || (b == '-' && idx > 0 && idx < lastIdx)) {
			l.emit("name-invalid", v)
			return
		}
	}

	if len(v) > 60 {
		l.emit("name-too-long", v)
	}
}

func (l *linter) lintDescription(v string) {
	if v == "" {
		l.emit("description-empty")
		return
	}

	if string(v[0]) != strings.ToUpper(string(v[0])) {
		l.emit("description-missing-uppercase", v)
	}

	if strings.HasSuffix(v, ".") {
		l.emit("description-extra-dot", v)
	}

	if len(v) > 120 {
		l.emit("description-too-long", v)
	}
}

func (l *linter) lintMaintainer(v string) {
	if v == "" {
		l.emit("maintainer-empty")
		return
	}

	_, err := mail.ParseAddress(v)
	if err != nil {
		l.emit("maintainer-invalid", v)
	}
}

func (l *linter) lintHomepage(v string) {
	if v == "" {
		l.emit("homepage-empty")
		return
	}

	url, err := url.Parse(v)
	if err != nil || url.Scheme == "" {
		l.emit("homepage-invalid", v)
	}
}

func (l *linter) lintSource(v *recipe.Source) {
	if v == nil {
		l.emit("source-empty")
		return
	}

	l.lintSourceURL(v.URL)
	l.lintSourceType(v.Type)
	l.lintSourceStrip(v.Strip)
}

func (l *linter) lintSourceURL(v string) {
	if v == "" {
		l.emit("source-url-empty")
		return
	}

	tmpl, err := template.New("").Parse(v)
	if err != nil {
		l.emit("source-url-invalid", v)
		return
	}

	buf := bytes.NewBuffer(nil)

	err = tmpl.Execute(buf, struct{ Version, Arch string }{"version", "arch"})
	if err != nil {
		l.emit("source-url-invalid", v)
		return
	}

	url, err := url.Parse(buf.String())
	if err != nil || url.Scheme == "" {
		l.emit("source-url-invalid", v)
	}
}

func (l *linter) lintSourceType(v string) {
	if v != "" && v != "archive" && v != "file" {
		l.emit("source-type-invalid", v)
	}
}

func (l *linter) lintSourceStrip(v int) {
	if v < 0 {
		l.emit("source-strip-invalid", v)
	}
}

func (l *linter) lintControl(v *recipe.Control) {
	if v == nil {
		l.emit("control-empty")
		return
	}

	l.lintControlDescription(v.Description)
}

func (l *linter) lintControlDescription(v string) {
	var max int

	if v == "" {
		l.emit("control-description-empty")
		return
	}

	for _, part := range strings.Split(v, "\n") {
		cur := len(part)
		if cur > deb.ControlDescriptionWrap {
			max = cur
		}
	}

	if max > 0 {
		l.emit("control-description-wrap", max)
	}
}

func (l *linter) lintInstall(v *recipe.Install) {
	if v == nil {
		l.emit("install-empty")
		return
	}

	l.lintInstallMap("recipe", v.Recipe)
	l.lintInstallMap("upstream", v.Upstream)
}

func (l *linter) lintInstallMap(subkey string, v recipe.InstallMap) {
	if subkey == "upstream" && v == nil {
		l.emit("install-upstream-empty")
		return
	}

	for dst, rules := range v {
		renames := make(map[string]struct{})

		if !filepath.IsAbs(dst) {
			l.emit("install-destination-relative", dst)
		}

		if rules == nil {
			l.emit("install-rule-empty", dst)
		}

		for idx, rule := range rules {
			if rule.Pattern == "" {
				l.emit("install-rule-pattern-empty", dst, idx)
			}

			if rule.Rename != "" {
				_, ok := renames[rule.Rename]
				if ok {
					l.emit("install-rule-rename-duplicate", dst, rule.Rename)
				} else {
					renames[rule.Rename] = struct{}{}
				}
			}

			if rule.ConfFile && !strings.HasPrefix(dst, "/etc") {
				l.emit("install-rule-conffile-outside-etc", dst, idx)
			}
		}
	}
}

func (l *linter) lintDirs(v []string) {
	for _, dir := range v {
		if !filepath.IsAbs(dir) {
			l.emit("dirs-path-relative", dir)
		}
	}
}

func (l *linter) lintLinks(v map[string]string) {
	for dst, src := range v {
		if !filepath.IsAbs(dst) {
			l.emit("links-destination-relative", dst)
		}
		if !filepath.IsAbs(src) {
			l.emit("links-source-relative", src)
		}
	}
}
