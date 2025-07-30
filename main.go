package main

import (
	"log"
	"os"

	"work-orchestrator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}