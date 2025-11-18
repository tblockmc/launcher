package tblock

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type LabelWriter struct {
	label *widget.Label
}

func (l *LabelWriter) Write(p []byte) (n int, err error) {
	fyne.Do(func() {
		text := string(p)
		if len(text) > 70 {
			text = text[:70] + "..."
		}

		l.label.SetText(text)
	})

	return len(p), nil
}
