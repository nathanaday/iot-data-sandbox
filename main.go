package main

import (
	"fmt"
	"log"

	"github.com/nathanaday/iot-data-sandbox/persistence"
)

func main() {

	store, err := persistence.NewStore("./iot-data.db")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

}
