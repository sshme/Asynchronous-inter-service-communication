package main

// @title           Orders Service API
// @version         1.0
// @description     A service for uploading and retrieving files
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost
// @BasePath  /orders-api

// @schemes http https
// @produce  json
// @consumes json multipart/form-data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"orders-service/internal/application/di"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "orders-service/docs"
)

func main() {
	app, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler: app.Router.SetupRoutes(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on port %d", app.Config.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
