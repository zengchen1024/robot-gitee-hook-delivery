package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/opensourceways/community-robot-lib/interrupts"
	"github.com/opensourceways/community-robot-lib/kafka"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	"github.com/opensourceways/community-robot-lib/mq"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
)

const component = "robot-gitee-hook-delivery"

type options struct {
	service        liboptions.ServiceOptions
	hmacSecretFile string
}

func (o *options) Validate() error {

	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.StringVar(&o.hmacSecretFile, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the HMAC secret.")

	_ = fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(component)

	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("invalid options")
	}

	cfg, err := loadConfig(o.service.ConfigFile)
	if err != nil {
		logrus.WithError(err).Fatal("load config")
	}

	// mq
	if err := initBroker(&cfg); err != nil {
		logrus.WithError(err).Fatal("Error init broker.")
	}

	defer func() {
		if err := kafka.Disconnect(); err != nil {
			logrus.Errorf("disconnet the mq failed, err:%s", err.Error())
		}
	}()

	// hmac
	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.hmacSecretFile}); err != nil {
		logrus.WithError(err).Fatal("start secret agent.")
	}

	defer secretAgent.Stop()

	hmac := secretAgent.GetTokenGenerator(o.hmacSecretFile)

	// server
	d := delivery{
		topic: cfg.Topic,
		hmac: func() string {
			return string(hmac())
		},
	}

	defer d.wait()

	run(&d, o.service.Port, o.service.GracePeriod)
}

func run(d *delivery, port int, gracePeriod time.Duration) {
	defer interrupts.WaitForGracefulShutdown()

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	// For /gitee-hook, handle a webhook normally.
	http.Handle("/gitee-hook", d)

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(port)}

	interrupts.ListenAndServe(httpServer, gracePeriod)
}

func initBroker(cfg *configuration) error {
	tlsConfig, err := cfg.Config.TLSConfig.TLSConfig()
	if err != nil {
		return err
	}

	err = kafka.Init(
		mq.Addresses(cfg.Config.Addresses...),
		mq.SetTLSConfig(tlsConfig),
		mq.Log(logrus.WithField("module", "broker")),
	)

	if err != nil {
		return err
	}

	return kafka.Connect()
}
