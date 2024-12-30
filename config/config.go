package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Telegram struct {
		PhoneNumber string `yaml:"phone_number"`
		ApiID       int    `yaml:"api_id"`
		ApiHash     string `yaml:"api_hash"`
		Password    string `yaml:"password"`
	} `yaml:"telegram"`

	Download struct {
		BaseDir string `yaml:"base_dir"`
	} `yaml:"download"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
