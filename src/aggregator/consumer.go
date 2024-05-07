package aggregator

import (
	"time"

	"github.com/sirupsen/logrus"

	lhm "github.com/xboshy/linkedhashmap"

	"p2p_aggregator/src/prom_metrics"

	"github.com/apache/pulsar-client-go/pulsar"
)

func P2p_aggregator(consume_chan <-chan pulsar.ConsumerMessage, write_chan chan<- []pulsar.ConsumerMessage, ack_chan chan<- pulsar.ConsumerMessage, time_delta int64, maxtotalinflightsize uint64, maxaggregatedsize uint64) {
	var n_read float64 = 0

	last_instant := time.Now()
	last_publish_time := time.Unix(0, 0)
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	mf := &mapFunctions{
		write_chan:     write_chan,
		total_size:     0,
		max_total_size: maxtotalinflightsize,
		time_delta:     time_delta,
	}

	agg := lhm.New[string, container](0, mf)

	for {
		select {
		case msg := <-consume_chan:
			n_read += 1
			prom_metrics.Prom_metric.Inc_number_read_msg()
			prom_metrics.Prom_metric.Inc_number_inflight_msg()

			cont := agg.Get(msg.Key())

			if cont == nil {
				cont = &container{
					min_publish_time: msg.PublishTime().Unix(),
					max_publish_time: msg.PublishTime().Unix(),
					agg_size:         0,
					messages:         make([]pulsar.ConsumerMessage, 0),
				}
			} else {
				if msg.PublishTime().Unix() > cont.max_publish_time {
					cont.max_publish_time = msg.PublishTime().Unix()
				}
				if msg.PublishTime().Unix() < cont.min_publish_time {
					cont.min_publish_time = msg.PublishTime().Unix()
				}
			}

			cont.messages = append(cont.messages, msg)
			cont.agg_size += uint64(len(msg.Payload()))
			mf.total_size += uint64(len(msg.Payload()))

			if cont.agg_size < maxaggregatedsize {
				agg.Push(msg.Key(), *cont)
			} else {
				agg.PullKey(msg.Key())
				k := msg.Key()
				mf.ExpiredHandler(&k, cont)

				// TODO metric pushed due to size
			}

			last_publish_time = msg.PublishTime()

			// // FIXME remove
			// ack_chan <- msg

		case <-tick.C:
			since := time.Since(last_instant)
			last_instant = time.Now()
			logrus.Infof("Read rate: %.3f msg/s; In flight agg %d; In flight size %d (last pulsar time %v)", n_read/float64(since/time.Second), agg.Len(), mf.total_size, last_publish_time)
			n_read = 0
		}
	}

}

func Acks(consumer pulsar.Consumer, ack_chan <-chan pulsar.ConsumerMessage) {
	last_instant := time.Now()
	last_publish_time := time.Unix(0, 0)

	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	var ack float64 = 0
	for {
		select {
		case msg := <-ack_chan:

			if err := consumer.Ack(msg); err != nil {
				logrus.Println(err)
			}
			ack++
			last_publish_time = msg.PublishTime()

		case <-tick.C:
			since := time.Since(last_instant)
			last_instant = time.Now()
			logrus.Infof("Ack rate: %.3f msg/s; (last pulsar time %v)", ack/float64(since/time.Second), last_publish_time)
			ack = 0
		}
	}
}
