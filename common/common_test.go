// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

package common

import (
	"testing"
	"unicode/utf8"
)

func TestTitle(t *testing.T) {
	txt := "This is my title"
	title := Title(txt)

	titleLength := utf8.RuneCountInString(title)
	lenTxt := utf8.RuneCountInString(txt)
	targetLen := 3*lenTxt + 6*2 + 2
	if titleLength != targetLen {
		t.Errorf("Title's length is not correct (%d/%d)", titleLength, targetLen)
	}
}

func TestEmail(t *testing.T) {
	email := "user@domain.fr"
	if !IsValidEmail(email) {
		t.Errorf("Email %s should be valid", email)
	}
	email = "userdomain"
	if IsValidEmail(email) {
		t.Errorf("Email %s should not be valid", email)
	}
}

func TestTranslation(t *testing.T) {
	msg := "app.title"
	translated := T(msg)
	if translated == msg {
		t.Errorf("Translation for %s should not be the same", msg)
	}
}
