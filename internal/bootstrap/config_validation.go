package bootstrap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"boilerplate-golang/internal/config"
)

func validateAppConfig(app config.App) error {
	if err := validateGinMode(app.GinMode); err != nil {
		return err
	}
	if err := validatePort("app.port", app.Port); err != nil {
		return err
	}
	if len(app.CORSAllowedOrigins) == 0 {
		return fmt.Errorf("app.cors_allowed_origins is required")
	}
	for _, origin := range app.CORSAllowedOrigins {
		if origin == "*" {
			return fmt.Errorf("app.cors_allowed_origins must not contain wildcard origin")
		}
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" {
			return fmt.Errorf("app.cors_allowed_origins contains invalid origin: %s", origin)
		}
	}
	return nil
}

func validateGinMode(mode string) error {
	switch mode {
	case "debug", "release", "test":
		return nil
	default:
		return fmt.Errorf("app.gin_mode must be one of: debug, release, test")
	}
}

func validatePostgreSQLConfig(postgres config.PostgreSQL) error {
	required := map[string]string{
		"database.pgsql.db_name.host":     postgres.Host,
		"database.pgsql.db_name.port":     postgres.Port,
		"database.pgsql.db_name.username": postgres.Username,
		"database.pgsql.db_name.password": postgres.Password,
		"database.pgsql.db_name.dbname":   postgres.DBName,
	}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", key)
		}
	}
	return validatePort("database.pgsql.db_name.port", postgres.Port)
}

func validatePort(name, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%s must be a valid port", name)
	}
	return nil
}
