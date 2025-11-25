package static

import (
	"embed"
	_ "embed"
)

//go:embed options.txt
var OptionsTXT []byte

//go:embed servers.dat
var ServersDAT []byte

//go:embed tblauncher.png
var Background []byte

//go:embed minecraft.ttf
var Font []byte

//go:embed translations
var Translations embed.FS
