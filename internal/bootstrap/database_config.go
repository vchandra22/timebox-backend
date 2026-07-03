package bootstrap

import (
	"timebox-backend/internal/config"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func LoadDatabaseConfig(configReader *viper.Viper, log *zap.Logger) config.Database {
	postgres := config.PostgreSQL{
		Host:     configReader.GetString("database.pgsql.db_name.host"),
		Port:     configReader.GetString("database.pgsql.db_name.port"),
		Username: configReader.GetString("database.pgsql.db_name.username"),
		Password: configReader.GetString("database.pgsql.db_name.password"),
		DBName:   configReader.GetString("database.pgsql.db_name.dbname"),
	}

	if err := validatePostgreSQLConfig(postgres); err != nil {
		log.Fatal("Invalid database config", zap.Error(err))
	}

	db, err := PostgreSQLInit(postgres, log)
	if err != nil {
		log.Fatal("Error connecting to database", zap.Error(err))
	}

	postgres.Conn = db

	return config.Database{
		PostgreSQL: postgres,
	}
}
