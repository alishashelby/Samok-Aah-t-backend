package config

import (
	"strings"

	"github.com/spf13/viper"
)

type LogConfig struct {
	Logger struct {
		Level    string `mapstructure:"level"`
		LogsDir  string `mapstructure:"logs-dir"`
		LogsFile string `mapstructure:"logs-file"`
	} `mapstructure:"logger"`
}

func LoadLogConfig() (*LogConfig, error) {
	v := viper.New()
	v.SetConfigName("log-config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg LogConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
