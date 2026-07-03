package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"timebox-backend/internal/bootstrap"
	"timebox-backend/internal/handler"
	"timebox-backend/internal/repository"
	"timebox-backend/internal/repository/dbexecutor"
	"timebox-backend/internal/router"
	"timebox-backend/internal/service"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var restCommand = &cobra.Command{
	Use:   "serve",
	Short: "Run API Service",
	Run: func(_ *cobra.Command, args []string) {

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		configReader := viper.New()

		log := bootstrap.LoggerInit()
		defer func() { _ = log.Sync() }()

		log.Info("Running application")

		appConfig := bootstrap.LoadConfig(configReader, log)
		defer appConfig.Database.PostgreSQL.Conn.Close()
		appConfig.Redis.Conn = bootstrap.RedisInit(ctx, appConfig.Redis, log)
		defer appConfig.Redis.Conn.Close()

		dbExecutor := dbexecutor.New(log)
		repositories := repository.New(appConfig.Database.PostgreSQL, appConfig.Redis, dbExecutor)
		services := service.New(repositories, appConfig.Redis.Conn, appConfig.JWT)
		handlers := handler.New(services)
		r := router.NewRouter(handlers, log, appConfig.App.CORSAllowedOrigins, appConfig.App.GinMode)

		addr := fmt.Sprintf(":%s", appConfig.App.Port)
		server := &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}

		log.Info("starting http server",
			zap.String("port", appConfig.App.Port),
		)

		errCh := make(chan error, 1)
		go func() {
			errCh <- server.ListenAndServe()
		}()

		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("failed to start server", zap.Error(err))
			}
		case <-ctx.Done():
			log.Info("shutdown signal received")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal("failed to shutdown server", zap.Error(err))
		}
	},
}
