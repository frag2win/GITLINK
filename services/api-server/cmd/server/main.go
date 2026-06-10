// Package main is the entry point for the api-server.
//
// It initializes the Fiber HTTP framework, loads configuration from
// environment variables, connects to the SQLite database, registers
// routes and middleware, and starts listening on the configured port.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/database"
	"github.com/localrepo/api-server/internal/handlers"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/router"
	"github.com/localrepo/api-server/internal/socket"
)

func main() {
	// ---- Load configuration ----
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// ---- Initialise database ----
	db, err := database.New(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to initialise database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	// ---- Create Fiber app ----
	app := fiber.New(fiber.Config{
		AppName:      "LocalRepo API Server",
		ServerHeader: "LocalRepo",
	})

	// ---- Initialise Git Client ----
	gitClient := socket.NewGitClient(cfg.GitSocketPath, 30*time.Second)
	handlers.Init(gitClient, db)

	// ---- Register global middleware ----
	middleware.SetupCORS(app)
	middleware.SetupLogger(app)

	// ---- Register routes ----
	router.Setup(app, db, cfg)

	// ---- Graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down server…")
		if err := app.Shutdown(); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	// ---- Start listening ----
	addr := ":" + cfg.Port
	log.Printf("api-server listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
