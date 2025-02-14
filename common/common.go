// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

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

// Check if an email is valid
func IsValidEmail(mail string) bool {
	return strings.Contains(mail, "@")
}

// Confirm a question
func Confirm(question string) bool {
	var response string

	fmt.Printf("%s [y/n] ", question)
	fmt.Scanln(&response)

	return strings.ToLower(response) == "y"
}

// Ask a question and return the response
func Ask(question string) string {
	var response string

	fmt.Printf("%s : ", question)
	fmt.Scanln(&response)

	return response
}
