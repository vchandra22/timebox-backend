package bootstrap

import (
	"boilerplate-golang/internal/config"

	"github.com/spf13/viper"
)

func LoadRedisConfig(configReader *viper.Viper) config.Redis {
	return config.Redis{
		Host:     configReader.GetString("redis.host"),
		Port:     configReader.GetString("redis.port"),
		Username: configReader.GetString("redis.username"),
		Password: configReader.GetString("redis.password"),
		DBName:   configReader.GetString("redis.dbname"),
		DBIndex:  configReader.GetInt("redis.dbindex"),
	}
}
