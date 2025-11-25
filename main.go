package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/havrydotdev/tblock-launcher/internal/tblock"
	"github.com/havrydotdev/tblock-launcher/pkg/launcher"
	"github.com/havrydotdev/tblock-launcher/pkg/utils"
)

// TODO: remove console logging in release mode
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error: \n", r)
		}
	}()

	gameDir, err := utils.GetTblockFolderPath()
	if err != nil {
		log.Fatal("failed to determine game folder: ", err)
	}

	cfg, err := launcher.ReadPersistedConfig(gameDir)
	if err != nil {
		log.Println("Failed to read config file: ", err)
		cfg = launcher.NewConfig("", gameDir)
	}

	l, err := tblock.NewLauncher(cfg)
	if err != nil {
		log.Fatal("failed to start launcher: ", err)
	}
	defer func() {
		err := launcher.PersistConfig(gameDir, l.Config)
		if err != nil {
			log.Println("Failed to persist config: ", err)
		}
	}()

	l.Run()
}
