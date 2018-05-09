package columns

import (
	"fmt"
	"strings"
)

func Print(items []string, margin int) {
	itemsLen := len(items)

	max := 0
	for _, s := range items {
		l := len(s)
		if l > max {
			max = l
		}
	}

	width := ttyWidth()
	if width == -1 {
		fmt.Println(strings.Join(items, "\n"))
		return
	}

	columnLen := max + margin
	columns := width / columnLen
	rows := (itemsLen + columns - 1) / columns
	if itemsLen < columns {
		columnLen = width / itemsLen
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			idx := j*rows + i
			if idx >= itemsLen {
				goto stop
			}
			fmt.Print(items[idx], strings.Repeat(" ", columnLen-len(items[idx])))
		}
	stop:
		fmt.Print("\n")
	}
}
