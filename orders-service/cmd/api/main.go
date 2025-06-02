package main

import (
	"log"
	"orders-service/internal/application/di"
)

func main() {
	_, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
}
