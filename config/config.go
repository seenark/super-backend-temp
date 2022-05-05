package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Configuration struct {
	Environment string       `mapstructure:"ENVIRONMENT"`
	Mongo       MongoConfig  `mapstructure:"MONGO"`
	App         AppConfig    `mapstructure:"APP"`
	Google      GoogleConfig `mapstructure:"GOOGLE"`
	// REDISHOST   string
}

type AppConfig struct {
	Domain      string `mapstructure:"DOMAIN"`
	Port        int    `mapstructure:"PORT"`
	SecretKey   string `mapstructure:"SECRETKEY"`
	Refresh     string `mapstructure:"REFRESH_SECRET_KEY"`
	AllowOrigin string `mapstructure:"ALLOW_ORIGIN"`
}

type MongoConfig struct {
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
	URL      string `mapstructure:"URL"`
}

type GoogleConfig struct {
	ProjectID  string `mapstructure:"PROJECT_ID"`
	BucketName string `mapstructure:"BUCKET_NAME"`
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")
	// read config from ENV
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// read config
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func GetConfig() Configuration {
	initConfig()
	config := Configuration{}
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	return config
}
