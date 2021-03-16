package util

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URLs             []string `yaml:"URLs"`
	MinTimeout       int      `yaml:"MinTimeout"`
	MaxTimeout       int      `yaml:"MaxTimeout"`
	NumberOfRequests int      `yaml:"NumberOfRequests"`
}

func GetConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("path couldn't be empty")
	}

	var (
		file *os.File
		err  error
	)
	if file, err = os.Open(path); err != nil {
		return nil, err
	}

	config := &Config{}

	if err := config.parseConfig(file); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) parseConfig(reader io.Reader) error {
	if err := yaml.NewDecoder(reader).Decode(c); err != nil {
		return err
	}

	return nil
}
