package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/havrydotdev/tblock-launcher/internal/tblock"
	"github.com/havrydotdev/tblock-launcher/pkg/launcher"
	"github.com/havrydotdev/tblock-launcher/pkg/utils"
)

//go:embed options.txt
var options []byte

//go:embed servers.dat
var servers []byte

//go:embed tblauncher.png
var background []byte

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()

	gameDir, err := utils.GetTblockFolderPath()
	if err != nil {
		log.Fatal("Failed to determine game folder: ", err)
	}

	cfg, err := launcher.ReadPersistedConfig(gameDir)
	if err != nil {
		log.Println("Failed to read config file: ", err)
		cfg = launcher.NewConfig("", gameDir)
	}

	l := tblock.NewLauncher(cfg, background, options, servers)
	l.Run()

	err = launcher.PersistConfig(gameDir, l.Config)
	if err != nil {
		log.Println("Failed to persist config: ", err)
	}
}
