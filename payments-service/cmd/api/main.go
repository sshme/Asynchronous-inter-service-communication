package main

// @title           Payments Service API
// @version         1.0
// @description     A service for processing payments and managing user accounts with transactional inbox/outbox patterns
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost
// @BasePath  /payments-api

// @schemes http https
// @produce  json
// @consumes json multipart/form-data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payments-service/internal/application/di"
	"payments-service/internal/infrastructure/persistence/postgres"
	"syscall"
	"time"

	_ "payments-service/docs"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	runServer()
}

func runMigrations() {
	fmt.Println("Starting database migration for payments-service...")

	app, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application for migration: %v", err)
	}

	migrationsPath := "internal/infrastructure/persistence/postgres/migrations"
	if err := postgres.RunMigrations(app.DB, migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Payments service migrations completed successfully.")
}

func runServer() {
	app, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.InboxProcessor.RegisterHandler("order.created", app.PaymentsService.ProcessOrderCreated)

	log.Println("Starting inbox processor...")
	app.InboxProcessor.Start(ctx)

	log.Println("Starting outbox publisher...")
	app.OutboxPublisher.Start(ctx)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler: app.Router.SetupRoutes(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting HTTP server on port %d", app.Config.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	log.Println("Stopping inbox processor...")
	app.InboxProcessor.Stop()

	log.Println("Stopping outbox publisher...")
	app.OutboxPublisher.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := app.DB.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server exited gracefully")
}
