package config

import (
	"fmt"
	"log"

	"github.com/fogodev/openvvar"
)

type Config struct {
	Environment string `config:"environment;default=DEV;options=DEV, HOMO, PROD, UNK;description=Host environment (DEV, HOMO, PROD or UNK)."`
	Debug       bool   `config:"debug;default=false"`

	QueueHost       string `config:"queue-host;default=localhost"`
	QueuePort       string `config:"queue-port;default=27017"`
	QueueUser       string `config:"queue-user;default=admin"`
	QueuePassword   string `config:"queue-password;required"`
	QueueSelfSigned bool   `config:"queue-self-signed;default=0"`
	QueueName       string `config:"queue-name;required"`
	ExchangeName    string `config:"exchange-name;required"`
}

func (c *Config) QueueAddress(protocol string) string {
	return fmt.Sprintf("%v://%v:%v@%v:%v/", protocol, c.QueueUser, c.QueuePassword, c.QueueHost, c.QueuePort)
}

var globalConfig *Config = nil

func New() *Config {
	if globalConfig == nil {
		globalConfig = new(Config)
		if err := openvvar.Load(globalConfig); err != nil {
			log.Fatalf("An error occurred for bad config reasons: %v", err)
		}
	}

	return globalConfig
}

func Get() *Config {
	if globalConfig == nil {
		panic("Trying to get a nil config, you must use New function to instantiate configs before getting it")
	}
	return globalConfig
}
