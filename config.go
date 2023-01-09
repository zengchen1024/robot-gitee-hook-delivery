package main

import (
	"errors"
	"regexp"
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
	addresses := parseAddress(c.KafkaAddress)
	if len(addresses) == 0 {
		err = errors.New("no valid address for kafka")

		return
	}

	if err = kafka.ValidateConnectingAddress(addresses); err != nil {
		return
	}

	cfg.Addresses = addresses

	return
}

func parseAddress(addresses string) []string {
	var reIpPort = regexp.MustCompile(
		`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}:[1-9][0-9]*$`,
	)

	v := strings.Split(addresses, ",")
	r := make([]string, 0, len(v))
	for i := range v {
		if reIpPort.MatchString(v[i]) {
			r = append(r, v[i])
		}
	}

	return r
}

func loadConfig(path string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(path, &cfg); err == nil {
		err = cfg.validate()
	}

	return
}
