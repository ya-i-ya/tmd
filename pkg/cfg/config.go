package cfg

import (
	"errors"
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

	Logging struct {
		Filename   string `yaml:"filename"`
		MaxSize    int    `yaml:"max_size"`
		MaxAge     int    `yaml:"max_age"`
		MaxBackups int    `yaml:"max_backups"`
		Compress   bool   `yaml:"compress"`
		Level      string `yaml:"level"`
	} `yaml:"logging"`

	Fetching struct {
		DialogsLimit  int `yaml:"dialogs_limit"`
		MessagesLimit int `yaml:"messages_limit"`
	} `yaml:"fetching"`

	Database struct {
		Dialect  string `yaml:"dialect"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
}

func (cfg *Config) Validate() error {
	if cfg.Telegram.ApiID == 0 {
		return errors.New("telegram.api_id is required and must be non-zero")
	}
	if cfg.Telegram.ApiHash == "" {
		return errors.New("telegram.api_hash is required")
	}
	if cfg.Telegram.PhoneNumber == "" {
		return errors.New("telegram.phone_number is required")
	}
	return nil
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
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
