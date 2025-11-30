package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/havrydotdev/tblock-launcher/internal/tblock"
)

// TODO: remove console logging in release mode
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error: \n", r)
		}
	}()

	l, err := tblock.NewLauncher()
	if err != nil {
		log.Fatal("failed to start launcher: ", err)
	}
	defer func() {
		err := l.PersistConfig()
		if err != nil {
			log.Println("Failed to persist config: ", err)
		}
	}()

	l.Run()
}
