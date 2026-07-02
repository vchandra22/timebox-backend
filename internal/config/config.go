package config

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	App      App
	Database Database
	Redis    Redis
	JWT      JWT
	External External
}

type App struct {
	Name               string
	GinMode            string
	Host               string
	Port               string
	CORSAllowedOrigins []string
}

type Database struct {
	PostgreSQL PostgreSQL
}

type PostgreSQL struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	Conn     *sqlx.DB
}

type Redis struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	DBIndex  int
	Conn     *redis.Client
}

type JWT struct {
	Secret string
}

type External struct {
	AWS        AWS
	Cloudinary Cloudinary
	RESTClient RESTClient
}

type AWS struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
}

type Cloudinary struct {
	CloudName string
	APIKey    string
	APISecret string
}

type RESTClient struct {
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}
