package main

import (
	"github.com/jnovack/flag"
)

type opt struct {
	sourcepulsar                  string
	sourcetopic                   string
	sourcesubscription            string
	sourcename                    string
	sourcetrustcerts              string
	sourcecertfile                string
	sourcekeyfile                 string
	sourceallowinsecureconnection bool

	destpulsar                  string
	desttopic                   string
	destsubscription            string
	destname                    string
	desttrustcerts              string
	destcertfile                string
	destkeyfile                 string
	destallowinsecureconnection bool

	batchmaxpublishdelay uint
	batchmaxmessages     uint
	batchingmaxsize      uint

	timedelta            int64
	maxaggregatedsize    uint64
	maxtotalinflightsize uint64

	pprofon       bool
	pprofdir      string
	pprofduration uint

	prometheusport uint
	loglevel       string
}

func from_args() opt {

	var opt opt

	flag.StringVar(&opt.sourcepulsar, "source_pulsar", "pulsar://localhost:6650", "Source pulsar address")
	flag.StringVar(&opt.sourcetopic, "source_topic", "persistent://public/default/in", "Source topic names (seperated by ;)")
	flag.StringVar(&opt.sourcesubscription, "source_subscription", "p2p_aggregator", "Source subscription name")
	flag.StringVar(&opt.sourcename, "source_name", "aggregator_consumer", "Source consumer name")
	flag.StringVar(&opt.sourcetrustcerts, "source_trust_certs", "", "Path for source pem file, for ca.cert")
	flag.StringVar(&opt.sourcecertfile, "source_cert_file", "", "Path for source cert.pem file")
	flag.StringVar(&opt.sourcekeyfile, "source_key_file", "", "Path for source key-pk8.pem file")
	flag.BoolVar(&opt.sourceallowinsecureconnection, "source_allow_insecure_connection", false, "Source allow insecure connection")

	flag.StringVar(&opt.destpulsar, "dest_pulsar", "pulsar://localhost:6650", "Destination pulsar address")
	flag.StringVar(&opt.desttopic, "dest_topic", "persistent://public/default/out", "Destination topic name")
	flag.StringVar(&opt.destname, "dest_name", "aggregator_consumer", "Destination producer name")
	flag.StringVar(&opt.desttrustcerts, "dest_trust_certs", "", "Path for destination pem file, for ca.cert")
	flag.StringVar(&opt.destcertfile, "dest_cert_file", "", "Path for destination cert.pem file")
	flag.StringVar(&opt.destkeyfile, "dest_key_file", "", "Path for destination key-pk8.pem file")
	flag.BoolVar(&opt.destallowinsecureconnection, "dest_allow_insecure_connection", false, "Dest allow insecure connection")

	flag.UintVar(&opt.batchmaxpublishdelay, "batch_max_publish_delay", 300, "How long to wait for batching in milliseconds")
	flag.UintVar(&opt.batchmaxmessages, "batch_max_messages", 100, "Max batch messages")
	flag.UintVar(&opt.batchingmaxsize, "batch_max_size", 512*1024, "Max batch size in bytes")

	flag.Int64Var(&opt.timedelta, "time_delta", 10, "The delta in seconds between the first and last aggregation")
	flag.Uint64Var(&opt.maxaggregatedsize, "max_aggregated_size", 256*1024, "Max size in bytes on aggregation")
	flag.Uint64Var(&opt.maxtotalinflightsize, "max_total_inflight_size", 1024*1024*1024, "Max total messages size in bytes")

	flag.BoolVar(&opt.pprofon, "pprof_on", false, "Profoling on?")
	flag.StringVar(&opt.pprofdir, "pprof_dir", "./pprof", "Directory for pprof file")
	flag.UintVar(&opt.pprofduration, "pprof_duration", 60*4, "Number of seconds to run pprof")

	flag.UintVar(&opt.prometheusport, "prometheus_port", 7700, "Prometheous port")
	flag.StringVar(&opt.loglevel, "log_level", "info", "Logging level: panic - fatal - error - warn - info - debug - trace")

	flag.Parse()

	return opt

}
