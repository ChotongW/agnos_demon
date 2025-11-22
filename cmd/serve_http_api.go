package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"agnos_demo/internal/database"
	"agnos_demo/internal/helpers"
	"agnos_demo/internal/routes"
	"agnos_demo/internal/service"

	"github.com/spf13/cobra"
)

var serveUserAPICmd = &cobra.Command{
	Use:   "serve-user-http-api",
	Short: "Start User HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {

		logger, err := helpers.CreateLogger("", nil, nil, "")
		if err != nil {
			return err
		}

		// Connect DB with context
		ctx := context.Background()
		db, err := database.ConnectDB(ctx)
		if err != nil {
			logger.Fatalf("Failed to connect to database: %v", err)
			return err
		}
		defer db.Close()

		svc, err := service.NewService(
			logger,
			db,
			&service.ServiceOptions{},
		)
		if err != nil {
			return err
		}

		config, err := routes.InitConfig()
		if err != nil {
			return err
		}

		router := routes.NewRouter(svc)

		// Init Http Server
		HttpServer := http.Server{
			Addr:              fmt.Sprintf(":%d", config.Port),
			Handler:           router,
			ReadHeaderTimeout: 15 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		}

		// Waiting os signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			<-quit

			logger.Infof("Gracefully shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := HttpServer.Shutdown(ctx); err != nil {
				logger.Errorf("Server forced to shutdown: %v", err)
			}

			logger.Infof("Server exited properly")
			os.Exit(0)
		}()

		// Start Server
		logger.Infof("Serving HTTP API at http://127.0.0.1:%d", config.Port)
		if err := HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Infof("HTTP server listen and serves failed: %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveUserAPICmd)
}
