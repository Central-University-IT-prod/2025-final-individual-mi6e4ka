package config

import (
	"log"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type Config struct {
	DB     DBConfig     `mapstructure:"db" validate:"required"`
	HTTP   HTTPConfig   `mapstructure:"http" validate:"required"`
	S3     S3Config     `mapstructure:"s3" validate:"required"`
	Ollama OllamaConfig `mapstructure:"ollama" validate:"required"`
}

type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	DBName   string `mapstructure:"db_name" validate:"required"`
}

type HTTPConfig struct {
	Port int `mapstructure:"port" validate:"required"`
}

type S3Config struct {
	Endpoint string `mapstructure:"endpoint" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
}

type OllamaConfig struct {
	BaseURL string `mapstructure:"base_url" validate:"required"`
	Model   string `mapstructure:"model" validate:"required"`
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	validateConfig(&config)

	return &config
}

func validateConfig(config *Config) {
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
}
