package bootstrap

import (
	"strings"

	"timebox-backend/internal/config"

	"github.com/spf13/viper"
)

func LoadAppConfig(configReader *viper.Viper) config.App {
	return config.App{
		Name:               configReader.GetString("app.name"),
		GinMode:            configReader.GetString("app.gin_mode"),
		Host:               configReader.GetString("app.host"),
		Port:               configReader.GetString("app.port"),
		CORSAllowedOrigins: stringSliceConfig(configReader, "app.cors_allowed_origins"),
	}
}

func stringSliceConfig(configReader *viper.Viper, key string) []string {
	values := configReader.GetStringSlice(key)
	if len(values) > 0 {
		return trimStrings(values)
	}

	raw := strings.TrimSpace(configReader.GetString(key))
	if raw == "" {
		return nil
	}

	return trimStrings(strings.Split(raw, ","))
}

func trimStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}
