package bootstrap

import (
	"timebox-backend/internal/config"

	"github.com/spf13/viper"
)

func LoadExternalConfig(configReader *viper.Viper) config.External {
	return config.External{
		AWS: config.AWS{
			Region:          configReader.GetString("external.aws.region"),
			AccessKeyID:     configReader.GetString("external.aws.access_key_id"),
			SecretAccessKey: configReader.GetString("external.aws.secret_access_key"),
			S3Bucket:        configReader.GetString("external.aws.s3_bucket"),
		},
		Cloudinary: config.Cloudinary{
			CloudName: configReader.GetString("external.cloudinary.cloud_name"),
			APIKey:    configReader.GetString("external.cloudinary.api_key"),
			APISecret: configReader.GetString("external.cloudinary.api_secret"),
		},
		RESTClient: config.RESTClient{
			BaseURL:        configReader.GetString("external.rest_client.base_url"),
			APIKey:         configReader.GetString("external.rest_client.api_key"),
			TimeoutSeconds: configReader.GetInt("external.rest_client.timeout_seconds"),
		},
	}
}

func LoadJWTConfig(configReader *viper.Viper) config.JWT {
	return config.JWT{
		Secret: configReader.GetString("jwt.secret"),
	}
}
