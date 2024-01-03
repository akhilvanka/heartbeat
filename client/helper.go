package client

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Enabled bool   `yaml:"enabled"`
		Port    string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		URI string `yaml:"uri"`
	} `yaml:"database"`
}

func ReturnConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
