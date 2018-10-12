package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func createConfig() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigFile("/etc/calm.yaml")
	cfg.SetEnvPrefix("CALM")

	cfg.SetDefault(MEMORY_CONFIG_KEY, "0")
	cfg.SetDefault(CPU_CONFIG_KEY, "0")
	cfg.SetDefault(USER_KEY, "root")

	err := cfg.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read in config: %v", err)
	}
	return cfg, nil
}
