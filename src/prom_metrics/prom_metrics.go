package prom_metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sirupsen/logrus"
)

type Prom_metrics struct {
	read_messages     prometheus.Counter
	writen_messages   prometheus.Counter
	inflight_messages prometheus.Gauge
}

func (prom_metric *Prom_metrics) Inc_number_read_msg() {
	prom_metric.read_messages.Inc()
}

func (prom_metric *Prom_metrics) Inc_number_writen_msg() {
	prom_metric.writen_messages.Inc()
}

func (prom_metric *Prom_metrics) Inc_number_inflight_msg() {
	prom_metric.inflight_messages.Inc()
}

func (prom_metric *Prom_metrics) Sub_number_inflight_msg(n int) {
	prom_metric.inflight_messages.Sub(float64(n))
}

var Prom_metric *Prom_metrics

func Setup_prometheus(prometheusport uint) {

	reg := prometheus.NewRegistry()

	Prom_metric = &Prom_metrics{
		read_messages: promauto.NewCounter(prometheus.CounterOpts{
			Name: "read_messages",
			Help: "The total number of messages read from pulsar.",
		}),
		writen_messages: promauto.NewCounter(prometheus.CounterOpts{
			Name: "writen_messages",
			Help: "The total number of messages writen to pulsar.",
		}),
		inflight_messages: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "inflight_messages",
			Help: "The total number of messages inflight.",
		}),
	}

	reg.MustRegister(Prom_metric.read_messages)
	reg.MustRegister(Prom_metric.writen_messages)
	reg.MustRegister(Prom_metric.inflight_messages)

	if prometheusport > 0 {

		http.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				// Pass custom registry
				Registry: reg,
			},
		))

		logrus.Infof("metrics exposed at: localhost:%d/metrics", prometheusport)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", prometheusport), nil); err != nil {
			logrus.Errorf("setup prometheus: %+v", err)
		}
	}
}
