package columns

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ttyWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin

	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	parts := strings.Split(strings.Trim(string(out), "\n"), " ")
	width, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}

	return width
}
