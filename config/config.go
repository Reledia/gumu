package config

import (
	"os"
	"path/filepath"

	toml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var (
	k       = koanf.New(".")
	parser  = toml.Parser()
	home, _ = os.LookupEnv("HOME")
	path    = filepath.Join(home, ".config/gumu/config.toml")

	defaultConfig = Config{
		ConfigProton: ConfigProton{
			Repos: []string{"cachyos/proton-cachyos", "GloriousEggroll/proton-ge-custom"},
		},
		ConfigPrefix: ConfigPrefix{
			Prefixs: []string{""},
		},
	}
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

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0o755)
		return &defaultConfig, defaultConfig.Save()
	}

	if err := k.Load(file.Provider(path), parser); err != nil {
		return &config, err
	}
	k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf"})
	return &config, nil
}

func (c *Config) Save() error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	f.Truncate(0)
	kk := koanf.New(".")
	kk.Load(structs.Provider(c, "koanf"), nil)
	bytes, err := kk.Marshal(parser)
	if err != nil {
		return err
	}
	_, err = f.WriteAt(bytes, 0)
	return err
}
