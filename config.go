// config.go
package createon

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	DataDir     string `mapstructure:"data_dir"`
	TemplateDir string `mapstructure:"template_dir"`
	AssetsDir   string `mapstructure:"assets_dir"`

	Paywall struct {
		TestNet    bool    `mapstructure:"testnet"`
		DefaultBTC float64 `mapstructure:"default_btc"`
		DefaultXMR float64 `mapstructure:"default_xmr"`
		Timeout    string  `mapstructure:"timeout"`
	} `mapstructure:"paywall"`
}

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
