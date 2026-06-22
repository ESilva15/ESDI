// Package config is responsible for the configuration of the application
package config

// NOTE: actually make a package out of this if possible - lets try

import (
	"os"

	"gopkg.in/yaml.v3"
)

var instance *ESDICfg

type ESDICfg struct {
	DefaultSim    string `yaml:"default_sim"`
	DefaultLayout string `yaml:"default_layout"`
}

func (cfg *ESDICfg) loadConfiguration(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &instance)
	if err != nil {
		return err
	}

	return nil
}

func Setup(path string) error {
	instance = &ESDICfg{}

	err := instance.loadConfiguration(path)
	if err != nil {
		return err
	}

	return nil
}

func GetCfg() *ESDICfg {
	if instance == nil {
		panic("configuration instance wasn't initialized")
	}

	return instance
}
