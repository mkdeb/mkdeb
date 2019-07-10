package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

var outputPath string

func main() {
	flag.StringVar(&outputPath, "o", "", "output file path")
	flag.Parse()

	if outputPath == "" {
		log.Fatal(`missing "-o" mandatory option`)
	}

	set, err := loadRules()
	if err != nil {
		log.Fatalf("cannot load rules: %s", err)
	}

	w, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot create output file: %s", err)
	}
	defer w.Close()

	err = template.Must(template.New("").Parse(`// Generated by go generate; DO NOT EDIT.

package lint

var rules = map[string]*RuleInfo{
{{- range .Rules }}
	"{{ .Tag }}": {
		Tag:   "{{ .Tag }}",
		Level: {{ .Level }},
		Description: `+"`\n{{ .Description }}\n`"+`,
	},
{{- end }}
}
`,
	)).Execute(w, set)
	if err != nil {
		log.Fatalf("cannot execute template: %s", err)
	}
}

func loadRules() (*ruleSet, error) {
	files, err := ioutil.ReadDir("rules")
	if err != nil {
		return nil, err
	}

	set := &ruleSet{}
	for _, file := range files {
		var v *ruleSet

		if file.IsDir() {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join("rules", file.Name()))
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, &v)
		if err != nil {
			return nil, err
		}

		set.Rules = append(set.Rules, v.Rules...)
	}

	return set, nil
}

type ruleSet struct {
	Rules []*rule `yaml:"rules"`
}

type rule struct {
	Tag         string    `yaml:"tag"`
	Level       ruleLevel `yaml:"level"`
	Description ruleDesc  `yaml:"description"`
}

type ruleLevel string

func (l *ruleLevel) UnmarshalYAML(fn func(interface{}) error) error {
	var v string

	err := fn(&v)
	if err != nil {
		return err
	}

	*l = ruleLevel("Level" + strings.Title(v))

	return nil
}

type ruleDesc string

func (d *ruleDesc) UnmarshalYAML(fn func(interface{}) error) error {
	var v string

	err := fn(&v)
	if err != nil {
		return err
	}

	*d = ruleDesc(strings.TrimSpace(v))

	return nil
}
