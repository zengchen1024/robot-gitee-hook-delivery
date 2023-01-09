package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/kafka"
	"github.com/opensourceways/community-robot-lib/mq"
	"github.com/sirupsen/logrus"
)

type delivery struct {
	wg    sync.WaitGroup
	hmac  func() string
	topic string
}

func (d *delivery) wait() {
	d.wg.Wait()
}

func (d *delivery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, _, ok := giteeclient.ValidateWebhook(w, r, d.hmac)
	if !ok {
		return
	}

	fmt.Fprint(w, "Event received. Have a nice day.")

	d.publish(payload, r.Header, eventType, eventGUID)
}

func (d *delivery) publish(payload []byte, h http.Header, eventType, eventGUID string) {
	msg := mq.Message{
		Header: map[string]string{
			"content-type":      h.Get("content-type"),
			"X-Gitee-Event":     h.Get("X-Gitee-Event"),
			"X-Gitee-Timestamp": h.Get("X-Gitee-Timestamp"),
			"X-Gitee-Token":     h.Get("X-Gitee-Token"),
		},
		Body: payload,
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		l := logrus.WithFields(
			logrus.Fields{
				"event-type": eventType,
				"event-id":   eventGUID,
			},
		)

		if err := kafka.Publish(d.topic, &msg); err != nil {
			l.Errorf("failed to publish msg, err:%s", err.Error())
		} else {
			l.Infof("publish message to topic(%s) successfully", d.topic)
		}
	}()
}
