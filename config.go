// config.go
package createon

import (
	"github.com/spf13/viper"
)

// Config holds all application configuration settings loaded from a YAML file.
// It includes server settings, directory paths, and cryptocurrency payment configuration.
type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	DataDir     string `mapstructure:"data_dir"`
	TemplateDir string `mapstructure:"template_dir"`
	AssetsDir   string `mapstructure:"assets_dir"`

	Paywall struct {
		TestNet     bool    `mapstructure:"testnet"`
		DefaultBTC  float64 `mapstructure:"default_btc"`
		DefaultXMR  float64 `mapstructure:"default_xmr"`
		Timeout     string  `mapstructure:"timeout"`
		XMRHost     string  `mapstructure:"xmr_host"`
		XMRUser     string  `mapstructure:"xmr_user"`
		XMRPassword string  `mapstructure:"xmr_password"`
		BTCHost     string  `mapstructure:"btc_host"`
	} `mapstructure:"paywall"`
}

// LoadConfig reads and parses configuration from the specified YAML file path.
// It returns a Config struct populated with all settings from the file.
// Returns an error if the file cannot be read or parsed.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
