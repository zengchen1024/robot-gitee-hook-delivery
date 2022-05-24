package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/opensourceways/community-robot-lib/kafka"
	"github.com/opensourceways/community-robot-lib/mq"
	"github.com/sirupsen/logrus"
)

type courier struct {
	wg    sync.WaitGroup
	topic string
}

func (c *courier) wait() {
	c.wg.Wait()
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (c *courier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, _, ok := c.parsePayload(w, r)
	if !ok {
		return
	}

	l := logrus.WithFields(
		logrus.Fields{
			"event-type": eventType,
			"event-id":   eventGUID,
		},
	)

	if err := c.publish(payload, r.Header, l); err != nil {
		l.WithError(err).Error()
	}
}

func (c *courier) publish(payload []byte, h http.Header, l *logrus.Entry) error {
	header := map[string]string{
		"content-type":      h.Get("content-type"),
		"X-Gitee-Event":     h.Get("X-Gitee-Event"),
		"X-Gitee-Timestamp": h.Get("X-Gitee-Timestamp"),
		"X-Gitee-Token":     h.Get("X-Gitee-Token"),
		"User-Agent":        "Robot-Gitee-Access",
	}

	msg := mq.Message{
		Header: header,
		Body:   payload,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		if err := kafka.Publish(c.topic, &msg); err != nil {
			l.WithError(err).Error("failed to publish msg")
		} else {
			l.Info(fmt.Sprintf("publish message to %s topic success", c.topic))
		}
	}()

	return nil
}

func (c *courier) parsePayload(
	w http.ResponseWriter, r *http.Request,
) (eventType string, eventGUID string, payload []byte, status int, ok bool) {
	defer r.Body.Close()

	// Header checks: It must be a POST with an event type and a signature.
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		responseHTTPError(w, status, "405 Method not allowed")
		return
	}

	if v := r.Header.Get("content-type"); v != "application/json" {
		status = http.StatusBadRequest
		responseHTTPError(w, status, "400 Bad Request: Hook only accepts content-type: application/json")
		return
	}

	if eventType = r.Header.Get("X-Gitee-Event"); eventType == "" {
		status = http.StatusBadRequest
		responseHTTPError(w, status, "400 Bad Request: Missing X-Gitee-Event Header")
		return
	}

	if eventGUID = r.Header.Get("X-Gitee-Timestamp"); eventGUID == "" {
		status = http.StatusBadRequest
		responseHTTPError(w, status, "400 Bad Request: Missing X-Gitee-Timestamp Header")
		return
	}

	sig := r.Header.Get("X-Gitee-Token")
	if sig == "" {
		status = http.StatusForbidden
		responseHTTPError(w, status, "403 Forbidden: Missing X-Gitee-Token")
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status = http.StatusInternalServerError
		responseHTTPError(w, status, "500 Internal Server Error: Failed to read request body")
		return
	}

	status = http.StatusOK
	ok = true
	return
}

func responseHTTPError(w http.ResponseWriter, statusCode int, response string) {
	logrus.WithFields(
		logrus.Fields{
			"response":    response,
			"status-code": statusCode,
		},
	).Debug(response)

	http.Error(w, response, statusCode)
}
