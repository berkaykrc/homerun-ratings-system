package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/config"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/healthcheck"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/notification"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/accesslog"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/content"
	"github.com/go-ozzo/ozzo-routing/v2/cors"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

func main() {
	flag.Parse()

	// create root logger tagged with server version
	logger := log.New().With(context.Background(), "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// create notification storage and service
	storage := notification.NewInMemoryStorage(logger)
	notificationService := notification.NewService(storage, logger, *cfg)

	// start cleanup worker in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go notificationService.StartCleanupWorker(ctx)

	// build HTTP server
	address := fmt.Sprintf(":%d", cfg.ServerPort)
	hs := &http.Server{
		Addr:    address,
		Handler: buildHandler(logger, notificationService, Version),
	}

	// start the HTTP server with graceful shutdown
	go func() {
		defer cancel() // cancel the context to stop background workers when server shuts down
		routing.GracefulShutdown(hs, 30*time.Second, logger.Infof)
	}()
	logger.Infof("server %v is running at %v", Version, address)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		os.Exit(-1)
	}
}

// buildHandler sets up the HTTP routing and middleware stack.
func buildHandler(logger log.Logger, notificationService notification.Service, version string) http.Handler {
	router := routing.New()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		content.TypeNegotiator(content.JSON),
		cors.Handler(cors.AllowAll),
	)

	healthcheck.RegisterHandlers(router, version)
	notification.RegisterHandlers(router, notificationService, logger)

	return router
}
