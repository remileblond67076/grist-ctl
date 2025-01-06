// Common tools
package common

import (
	"fmt"
	"strings"
)

// Format string as a title
func Title(txt string) string {
	len := len(txt)
	line := strings.Repeat("-", len)
	newText := fmt.Sprintf("%s\n%s\n%s", line, txt, line)

	return newText
}

// Displays a title
func DisplayTitle(txt string) {
	fmt.Println(Title(txt))
}
