package config

import (
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const configPath = "config.yaml"

type DbConfig struct {
	User     string `yaml:"db_user"`
	Password string `yaml:"db_password"`
	Host     string `yaml:"db_host"`
	Port     string `yaml:"db_port"`
	DBName   string `yaml:"db_name"`
}

type AppConfig struct {
	DB       DbConfig `yaml:",inline"`
	HTTPHost string   `yaml:"http_host"`
	HTTPPort string   `yaml:"http_port"`
}

func ParseConfig() (*AppConfig, error) {
	config := &AppConfig{}

	confFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, xerrors.Errorf("Failed to read config file: %+v", err)
	}

	confFile = []byte(os.ExpandEnv(string(confFile)))

	if err = yaml.Unmarshal(confFile, config); err != nil {
		return nil, xerrors.Errorf("Cannot unmarshal config: %+v", err)
	}

	return config, nil
}
