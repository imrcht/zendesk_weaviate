package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	Openai struct {
		ApiKey string `mapstructure:"api_key"`
	}
	Embeddings struct {
		ApiType       string `mapstructure:"api_type"`
		Host          string `mapstructure:"host"`
		ApiKey        string `mapstructure:"api_key"`
		Model         string `mapstructure:"model"`
		Temperature   int    `mapstructure:"temperature"`
		ApiAuthHeader string `mapstructure:"apiAuthHeader"`
		ApiAuthToken  string `mapstructure:"apiAuthToken"`
	} `mapstructure:"embeddings"`
	Misc struct {
		EncryptedKey  string `mapstructure:"encryptedKey"`
		RunType       string `mapstructure:"run_type"`
		JsonDirectory string `mapstructure:"json_directory"`
		TeamId        int    `mapstructure:"team_id"`
		LimitContacts int    `mapstructure:"limit_contacts"`
		NumWorkers    int    `mapstructure:"numWorkers"`
	} `mapstructure:"misc"`
	Weaviate struct {
		weaviateHost   string `mapstructure:"localhost:8090"`
		weaviateScheme string `mapstructure:"http"`
	} `mapstructure:"weaviate"`
	Zendesk struct {
		Subdomain    string `mapstructure:"subdomain"`
		AccesssToken string `mapstructure:"access_token"`
	}
}

var (
	config *Config
)

func LoadConfig(path string) error {
	log.Println("Loading config...")

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Failed to read configuration file:", err)
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Println("Failed to unmarshal configuration:", err)
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	fmt.Println("Config loaded.")
	return nil
}

func GetConfig() *Config {
	return config
}
