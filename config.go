package main

import (
	"errors"

	"github.com/opensourceways/community-robot-lib/mq"
	"github.com/opensourceways/community-robot-lib/utils"
)

type configuration struct {
	Config mq.MQConfig `json:"config" required:"true"`
	Topic  string      `json:"topic"  required:"true"`
}

func (c *configuration) validate() error {
	if c.Topic == "" {
		return errors.New("missing topic")
	}

	return nil
}

func loadConfig(path string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(path, &cfg); err == nil {
		err = cfg.validate()
	}

	return
}
