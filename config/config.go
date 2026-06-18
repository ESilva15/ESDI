// package config is responsible for the configuration of the application
package config

// NOTE: actually make a package out of this if possible - lets try

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	instance *ESDICfg
	once     sync.Once
)

type ESDICfg struct{}

func (cfg *ESDICfg) loadConfiguration() error {
	file, err := os.ReadFile("./cfg.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &instance)
	if err != nil {
		return err
	}

	return nil
}

func Setup() error {
	instance = &ESDICfg{}

	err := instance.loadConfiguration()
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
