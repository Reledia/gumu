package config

import (
	"os"
	"path/filepath"

	toml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
)

var (
	k       = koanf.New(".")
	parser  = toml.Parser()
	home, _ = os.LookupEnv("HOME")
	path    = filepath.Join(home, ".config/gumu/config.toml")
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
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o600)

		log.Err(err).Send()

		config = Config{
			ConfigProton: ConfigProton{
				Repos: []string{"cachyos/proton-cachyos"},
			},
			ConfigPrefix: ConfigPrefix{
				Prefixs: []string{""},
			},
		}

		err = k.Load(structs.Provider(config, "koanf"), nil)
		if err != nil {
			log.Error().Err(err).Send()
			return &config, err
		}
		bytes, err := k.Marshal(parser)
		if err != nil {
			log.Error().Err(err).Send()
			return &config, err
		}
		_, err = f.Write(bytes)
		if err != nil {
			log.Error().Err(err).Send()
			return &config, err
		}
	} else {
		if err := k.Load(file.Provider(path), parser); err != nil {
			return &config, err
		}
		k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf"})
	}
	return &config, nil
}

func (c *Config) Save() error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0o755)
	if err != nil {
		log.Error().Err(err).Send()
	}
	f.Truncate(0)
	kk := koanf.New(".")
	kk.Load(structs.Provider(c, "koanf"), nil)
	bytes, err := kk.Marshal(parser)
	if err != nil {
		log.Error().Err(err).Send()
	}
	_, err = f.WriteAt(bytes, 0)
	if err != nil {
		log.Error().Err(err).Send()
	}
	return err
}
