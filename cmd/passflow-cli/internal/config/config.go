package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIURL           string `mapstructure:"api_url" yaml:"api_url"`
	DefaultWorkspace string `mapstructure:"default_workspace" yaml:"default_workspace"`
	Token            string `mapstructure:"token" yaml:"token"`
}

var (
	cfg        Config
	configFile string
)

func SetConfigFile(file string) {
	configFile = file
}

func Load() error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}

		configDir := filepath.Join(home, ".passflow")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("cannot create config directory: %w", err)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetDefault("api_url", "https://api.passflow.ai")
	viper.SetDefault("default_workspace", "")
	viper.SetDefault("token", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return viper.Unmarshal(&cfg)
}

func Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".passflow", "config.yaml")
	return viper.WriteConfigAs(configPath)
}

func GetAPIURL() string {
	return viper.GetString("api_url")
}

func SetAPIURL(url string) error {
	viper.Set("api_url", url)
	cfg.APIURL = url
	return Save()
}

func GetWorkspace() string {
	return viper.GetString("default_workspace")
}

func SetWorkspace(workspace string) error {
	viper.Set("default_workspace", workspace)
	cfg.DefaultWorkspace = workspace
	return Save()
}

func GetToken() string {
	return viper.GetString("token")
}

func SetToken(token string) error {
	viper.Set("token", token)
	cfg.Token = token
	return Save()
}

func ClearToken() error {
	return SetToken("")
}

func GetAll() Config {
	return Config{
		APIURL:           GetAPIURL(),
		DefaultWorkspace: GetWorkspace(),
		Token:            GetToken(),
	}
}

func IsAuthenticated() bool {
	return GetToken() != ""
}
