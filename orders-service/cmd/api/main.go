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
	"orders-service/internal/infrastructure/config"
	"orders-service/internal/infrastructure/persistence/postgres"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "orders-service/docs"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	runServer()
}

func runMigrations() {
	fmt.Println("Starting database migration...")
	// We need to load config just for the database
	cfg := config.MustLoad(config.NewApp("config/config.yaml"))
	dbConf := di.NewPostgresConfig(cfg)
	db, err := postgres.NewDb(dbConf)
	if err != nil {
		log.Fatalf("failed to connect to database for migration: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	fmt.Println("Migrations completed successfully.")
}

func runServer() {
	app, cleanup, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.InboxProcessor.RegisterHandler("payment.completed", app.OrdersService.ProcessPaymentCompleted)
	app.InboxProcessor.RegisterHandler("payment.failed", app.OrdersService.ProcessPaymentFailed)

	app.OutboxPublisher.Start(ctx)
	defer app.OutboxPublisher.Stop()

	app.InboxProcessor.Start(ctx)
	defer app.InboxProcessor.Stop()

	app.SSEManager.Start(ctx)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler: app.Router.SetupRoutes(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting Orders Service on port %d", app.Config.Server.Port)
		log.Printf("Outbox Publisher started for event publishing")
		log.Printf("Inbox Processor started for payment event handling")
		log.Printf("SSE Manager started for real-time order status updates")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down Orders Service...")

	app.InboxProcessor.Stop()
	app.OutboxPublisher.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Orders Service exited gracefully")
}
