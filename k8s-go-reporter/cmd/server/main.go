package main

import (
	"log"

	"github.com/carlosealves2/k8s-go-reporter/internal/bootstrap"
)

var (
	VERSION = "v0.0.1"
	COMMIT  = "#abcd"
)

func main() {
	// Create dependency injection container
	container, err := bootstrap.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Run application with injected dependencies
	if err := bootstrap.Run(container, VERSION, COMMIT); err != nil {
		log.Fatal(err)
	}
}
