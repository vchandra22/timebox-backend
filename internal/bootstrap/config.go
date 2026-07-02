package bootstrap

import (
	"boilerplate-golang/internal/config"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func LoadConfig(configReader *viper.Viper, log *zap.Logger) *config.Config {
	LoadConfigFile(configReader, log)

	app := LoadAppConfig(configReader)
	if err := validateAppConfig(app); err != nil {
		log.Fatal("Invalid app config", zap.Error(err))
	}

	db := LoadDatabaseConfig(configReader, log)
	redis := LoadRedisConfig(configReader)
	jwt := LoadJWTConfig(configReader)
	external := LoadExternalConfig(configReader)

	return &config.Config{
		App:      app,
		Database: db,
		Redis:    redis,
		JWT:      jwt,
		External: external,
	}
}

func LoadConfigFile(configReader *viper.Viper, log *zap.Logger) {
	configReader.SetConfigName("config")
	configReader.SetConfigType("yaml")
	configReader.AddConfigPath(".")

	if err := configReader.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file", zap.Error(err))
	}
}
