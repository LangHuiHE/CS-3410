package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func setup() error {
	InitCommands()
	err := SetupWorld()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := setup()
	if err != nil {
		log.Fatalf("Setting up: %v", err)
	}

	// connection part:
	inputChannel := make(chan InputEvent)
	go ListenForConnections(inputChannel)
	MainLoop(inputChannel)
}
