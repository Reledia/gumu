package config

import (
	toml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	k      = koanf.New(".")
	parser = toml.Parser()
)

type ConfigProton struct {
	Repos []string `koanf:"repos"`
}

type ConfigPrefix struct {
	Prefixs []string `koanf:"prefixs"`
}

type Config struct {
	ConfigProton `koanf:"proton"`
	ConfigPrefix `koanf:"prefix"`
}

func LoadConf() (*Config, error) {
	var config Config

	if err := k.Load(file.Provider("../config/test.toml"), parser); err != nil {
		return &config, err
	}

	k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf"})
	return &config, nil
}
