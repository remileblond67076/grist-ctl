package common

import (
	"testing"
)

func TestTitle(t *testing.T) {
	txt := "Ceci est mon texte"
	title := Title(txt)

	lenTitle := len(title)
	lenTxt := len(txt)
	if lenTitle != (3*lenTxt + 2) {
		t.Errorf("La longueur du titre n'est pas correcte (%d/%d)", lenTitle, lenTxt)
	}
}
