package main

import (
	"strings"
	"time"

	"p2p_aggregator/src/aggregator"
	"p2p_aggregator/src/prom_metrics"

	"github.com/sirupsen/logrus"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
)

func logging(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	l, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Fatalln("Failed parse log level. Reason: ", err)
	}
	logrus.SetLevel(l)
}

func new_client(url string, trust_cert_file string, cert_file string, key_file string, allow_insecure_connection bool) pulsar.Client {
	var client pulsar.Client
	var err error
	var auth pulsar.Authentication

	if len(cert_file) > 0 || len(key_file) > 0 {
		auth = pulsar.NewAuthenticationTLS(cert_file, key_file)
	}

	client, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:                        url,
		TLSAllowInsecureConnection: allow_insecure_connection,
		Authentication:             auth,
		TLSTrustCertsFilePath:      trust_cert_file,
		Logger:                     log.NewLoggerWithLogrus(logrus.New()),
	})

	if err != nil {
		logrus.Errorf("Failed connect to pulsar. Reason: %+v", err)
	}
	return client
}

func main() {
	opt := from_args()
	logging(opt.loglevel)
	logrus.Infof("%+v", opt)

	go prom_metrics.Setup_prometheus(opt.prometheusport)

	source_client := new_client(opt.sourcepulsar, opt.sourcetrustcerts, opt.sourcecertfile, opt.sourcekeyfile, opt.sourceallowinsecureconnection)
	dest_client := new_client(opt.destpulsar, opt.desttrustcerts, opt.destcertfile, opt.destkeyfile, opt.destallowinsecureconnection)

	defer source_client.Close()
	defer dest_client.Close()

	consume_chan := make(chan pulsar.ConsumerMessage, 2000)
	write_chan := make(chan []pulsar.ConsumerMessage, 1000)
	ack_chan := make(chan pulsar.ConsumerMessage, 2000)

	// Have 2 modes partioned topics or multiple topcis

	consumer, err := source_client.Subscribe(pulsar.ConsumerOptions{
		Topics:                      strings.Split(opt.sourcetopic, ";"),
		SubscriptionName:            opt.sourcesubscription,
		Name:                        opt.sourcename,
		Type:                        pulsar.Failover,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
		MessageChannel:              consume_chan,
		ReceiverQueueSize:           2000,
		NackRedeliveryDelay:         time.Duration(opt.timedelta * 4 * int64(time.Second)),
	})
	if err != nil {
		logrus.Fatal("Failed create consumer. Reason: ", err)
	}

	producer, err := dest_client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   opt.desttopic,
		Name:                    opt.destname,
		BatchingMaxPublishDelay: time.Millisecond * time.Duration(opt.batchmaxpublishdelay),
		BatchingMaxMessages:     opt.batchmaxmessages,
		BatchingMaxSize:         opt.batchingmaxsize,
	})
	if err != nil {
		logrus.Fatal("Failed create producer. Reason: ", err)
	}

	defer consumer.Close()
	defer producer.Close()

	if opt.pprofon {
		go activate_profiling(opt.pprofdir, time.Duration(opt.pprofduration)*time.Second)
	}

	go aggregator.P2p_aggregator(consume_chan, write_chan, ack_chan, opt.timedelta, opt.maxtotalinflightsize, opt.maxaggregatedsize)
	go aggregator.Writer(producer, write_chan, ack_chan, opt.batchingmaxsize)

	aggregator.Acks(consumer, ack_chan)
}
