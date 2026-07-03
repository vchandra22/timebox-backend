package cmd

import (
	"context"
	"timebox-backend/internal/bootstrap"
	"timebox-backend/internal/migration"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var migrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(_ *cobra.Command, args []string) {
		ctx := context.Background()
		configReader := viper.New()

		log := bootstrap.LoggerInit()
		bootstrap.LoadConfigFile(configReader, log)
		dbConfig := bootstrap.LoadDatabaseConfig(configReader, log)
		defer dbConfig.PostgreSQL.Conn.Close()

		applied, err := migration.Run(ctx, dbConfig.PostgreSQL.Conn, "migrations")
		if err != nil {
			log.Fatal("failed to run migrations", zap.Error(err))
		}

		log.Info("database migrations completed", zap.Int("applied", len(applied)), zap.Strings("files", applied))
	},
}
