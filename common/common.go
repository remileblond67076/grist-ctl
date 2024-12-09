package common

import (
	"fmt"
	"strings"
)

func Title(txt string) string {
	len := len(txt)
	line := strings.Repeat("-", len)
	newText := fmt.Sprintf("%s\n%s\n%s", line, txt, line)

	return newText
}

func DisplayTitle(txt string) {
	fmt.Println(Title(txt))
}
