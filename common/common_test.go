// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

package common

import (
	"testing"
)

func TestTitle(t *testing.T) {
	txt := "This is my title"
	title := Title(txt)

	titleLength := len(title)
	lenTxt := len(txt)
	if titleLength != (3*lenTxt + 2) {
		t.Errorf("Title's length is not correct (%d/%d)", titleLength, lenTxt)
	}
}
