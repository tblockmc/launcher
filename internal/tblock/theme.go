package tblock

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/havrydotdev/tblock-launcher/internal/static"
)

type tblockTheme struct{}

func newTheme() fyne.Theme {
	return &tblockTheme{}
}

func (t tblockTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameForeground {
		return color.RGBA{103, 176, 12, 255}
	}

	if name == theme.ColorNameFocus {
		return color.Transparent
	}

	return theme.DarkTheme().Color(name, variant)
}

func (t tblockTheme) Font(fyne.TextStyle) fyne.Resource {
	return fyne.NewStaticResource("minecraft_font", static.Font)
}

func (t tblockTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

func (t tblockTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(name)
}
