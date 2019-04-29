package print

import (
	"fmt"
	"os"

	"github.com/mgutz/ansi"
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
