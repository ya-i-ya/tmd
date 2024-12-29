package config

type Config struct {
	Telegram struct {
		PhoneNumber string `yaml:"phone_number"`
		ApiID       int    `yaml:"api_id"`
		ApiHash     string `yaml:"api_hash"`
	} `yaml:"telegram"`
	Download struct {
		BaseDir string `yaml:"base_dir"`
	} `yaml:"download"`
}
