package main

import (
	"errors"
	"strings"

	"github.com/opensourceways/community-robot-lib/kafka"
	"github.com/opensourceways/community-robot-lib/mq"
	"github.com/opensourceways/community-robot-lib/utils"
)

type configuration struct {
	Topic        string `json:"topic"          required:"true"`
	UserAgent    string `json:"user_agent"     required:"true"`
	KafkaAddress string `json:"kafka_address"  required:"true"`
}

func (c *configuration) validate() error {
	if c.Topic == "" {
		return errors.New("missing topic")
	}

	if c.UserAgent == "" {
		return errors.New("missing user_agent")
	}

	if c.KafkaAddress == "" {
		return errors.New("missing kafka_address")
	}

	return nil
}

func (c *configuration) kafkaConfig() (cfg mq.MQConfig, err error) {
	v := strings.Split(c.KafkaAddress, ",")
	if err = kafka.ValidateConnectingAddress(v); err == nil {
		cfg.Addresses = v
	}

	return
}

func loadConfig(path string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(path, &cfg); err == nil {
		err = cfg.validate()
	}

	return
}
