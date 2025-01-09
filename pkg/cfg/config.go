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

	Minio struct {
		Endpoint  string `yaml:"endpoint"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
		Bucket    string `yaml:"bucket"`
		BasePath  string `yaml:"base_path"`
		UseSSL    bool   `yaml:"use_ssl"`
	} `yaml:"minio"`
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
	if cfg.Database.Dialect == "" {
		return errors.New("database.dialect is required")
	}
	if cfg.Database.Host == "" {
		return errors.New("database.host is required")
	}
	if cfg.Database.Port == 0 {
		return errors.New("database.port is required")
	}
	if cfg.Database.DBName == "" {
		return errors.New("database.dbname is required")
	}
	if cfg.Minio.Endpoint == "" {
		return errors.New("minio.endpoint is required")
	}
	if cfg.Minio.AccessKey == "" || cfg.Minio.SecretKey == "" {
		return errors.New("minio.access_key and minio.secret_key are required")
	}
	if cfg.Minio.Bucket == "" {
		return errors.New("minio.bucket is required")
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
