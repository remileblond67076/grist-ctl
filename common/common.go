// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Common tools
package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/Xuanwo/go-locale"
	"github.com/mattn/go-colorable"
	"github.com/muesli/termenv"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var localizer *i18n.Localizer // Global localizer
var bundle *i18n.Bundle       // Global bundle

func init() {
	// Detect the language
	tag, err := locale.Detect()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize i18n with English (default) and French languages
	bundle = i18n.NewBundle(language.English)            // Default language
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal) // Register JSON unmarshal function
	bundle.LoadMessageFile("locales/en.json")            // Load English messages
	bundle.LoadMessageFile("locales/fr.json")            // Load French messages

	localizer = i18n.NewLocalizer(bundle, language.Tag.String(tag)) // Initialize localizer with detected language
}

// Translate a message
func T(msg string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: msg})
}

// Format string as a title
func Title(txt string) string {
	len := utf8.RuneCountInString(txt)
	line := strings.Repeat("═", len+2)
	newText := fmt.Sprintf("╔%s╗\n║ %s ║\n╚%s╝", line, txt, line)

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

	fmt.Printf("%s [%s/%s] ", question, T("questions.y"), T("questions.n"))
	fmt.Scanln(&response)

	return strings.ToLower(response) == T("questions.y")
}

// Ask a question and return the response
func Ask(question string) string {
	var response string

	fmt.Printf("%s : ", question)
	fmt.Scanln(&response)

	return response
}

// Print an example command line
func PrintCommand(txt string) {
	stdout := colorable.NewColorableStdout()

	profile := termenv.ColorProfile()

	if profile != termenv.Ascii {
		cmdText := termenv.String(txt).
			Foreground(termenv.ANSIRed).
			Background(termenv.ANSIWhite).
			String()
		fmt.Fprint(stdout, cmdText)
	} else {
		fmt.Print(txt)
	}
}

// Open a browser Window on an URL
func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
