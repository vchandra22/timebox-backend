package bootstrap

import (
	"net"
	"net/url"

	"timebox-backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func PostgreSQLInit(postgres config.PostgreSQL, log *zap.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", PostgreSQLDSN(postgres))
	if err != nil {
		log.Error("Failed opening database connection",
			zap.String("dbname", postgres.DBName),
			zap.Error(err),
		)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(DBMaxPool)
	db.SetMaxIdleConns(DBMaxIdle)
	db.SetConnMaxLifetime(DBMaxLifeTime)

	log.Info("Database connected", zap.String("dbname", postgres.DBName))

	return db, nil
}

func PostgreSQLDSN(postgres config.PostgreSQL) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(postgres.Username, postgres.Password),
		Host:   net.JoinHostPort(postgres.Host, postgres.Port),
		Path:   postgres.DBName,
	}
	q := dsn.Query()
	q.Set("sslmode", "disable")
	dsn.RawQuery = q.Encode()

	return dsn.String()
}
