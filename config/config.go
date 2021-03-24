package config

import (
	"fmt"

	"github.com/sherifabdlnaby/configuro"
)

type (
	Config struct {
		Api     *ApiConfig
		Slack   *SlackConfig
		Logger  *LoggerConfig
		Elastic *ElasticConfig
	}

	ApiConfig struct {
		Host string
		Port int
	}

	SlackConfig struct {
		Debug             bool
		BotID             string `config:"bot_id"`
		AccessToken       string `config:"access_token"`
		VerificationToken string `config:"verification_token"`
		SigningSecret     string `config:"signing_secret"`
	}

	LoggerConfig struct {
		Level  string
		Format string
	}

	ElasticConfig struct {
		Host string
		Port int
	}
)

func (c *ApiConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *ElasticConfig) DSN() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

func NewConfig(configPath string) *Config {
	conf, err := configuro.NewConfig(configuro.WithLoadFromConfigFile(configPath, true))
	if err != nil {
		panic(err)
	}

	confStruct := &Config{}
	if err := conf.Load(confStruct); err != nil {
		panic(err)
	}

	return confStruct
}
