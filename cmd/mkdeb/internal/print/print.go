package print

import (
	"fmt"
	"os"
	"strings"

	"github.com/mgutz/ansi"
	"mkdeb.sh/lint"
	"mkdeb.sh/recipe"
)

var enableEmoji = true

// DisableEmoji disables emojis printing.
func DisableEmoji() {
	enableEmoji = false
}

// Error prints an error message.
func Error(s string, args ...interface{}) {
	fmt.Fprint(os.Stderr, ansi.Color("Error: ", "red"))
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

// Lint prints linting problems.
func Lint(rcp *recipe.Recipe, problems []*lint.Problem) {
	var level string

	if problems == nil {
		return
	}

	for _, p := range problems {
		if p.Level == lint.LevelError {
			level = ansi.Color("E:", "red")
		} else {
			level = ansi.Color("W:", "yellow")
		}

		args := []string{""}
		for _, arg := range p.Args {
			args = append(args, fmt.Sprintf("%q", arg))
		}

		fmt.Printf(
			"%s %s: %s%s\n",
			level,
			rcp.Name,
			p.Tag,
			strings.Join(args, " "),
		)
	}
}

// LintInfo prints linting rule information.
func LintInfo(info *lint.RuleInfo) {
	var level string

	if info.Level == lint.LevelError {
		level = ansi.Color("error", "red")
	} else {
		level = ansi.Color("warning", "yellow")
	}

	fmt.Println(ansi.Color(info.Tag, "default+u"))
	fmt.Println("   " + strings.ReplaceAll(info.Description, "\n", "\n   "))
	fmt.Printf("   Level: %s\n\n", level)
}

// Section prints a section message.
func Section(s string, args ...interface{}) {
	print("green", s, args...)
}

// Step prints a step message.
func Step(s string, args ...interface{}) {
	print("blue", s, args...)
}

// Summary prints a summary message.
func Summary(emoji, s string, args ...interface{}) {
	Step("Summary")
	if enableEmoji {
		fmt.Print(emoji + "  ")
	}
	fmt.Printf(s+"\n", args...)
}

func print(color, s string, args ...interface{}) {
	fmt.Print(ansi.Color("==> ", color))
	fmt.Printf(ansi.Color(s, "default+b")+"\n", args...)
}
