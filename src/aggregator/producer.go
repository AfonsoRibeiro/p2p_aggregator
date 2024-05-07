package aggregator

import (
	"context"
	"p2p_aggregator/src/prom_metrics"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
)

func write_callback(agg []pulsar.ConsumerMessage, ack_chan chan<- pulsar.ConsumerMessage) func(msgID pulsar.MessageID, pm *pulsar.ProducerMessage, err error) {
	return func(msgID pulsar.MessageID, pm *pulsar.ProducerMessage, err error) {
		if err == nil {
			// Inflight_messages
			for i := range agg {
				// FIXME uncomment
				ack_chan <- agg[i]
			}
		} else {
			//FIXME nack
			// TODO metric nack
		}
		prom_metrics.Prom_metric.Inc_number_writen_msg()
	}
}

func Writer(producer pulsar.Producer, write_chan <-chan []pulsar.ConsumerMessage, ack_chan chan<- pulsar.ConsumerMessage, batchingmaxsize uint) {
	var n_writen float64 = 0
	var n_ignored uint64 = 0

	last_instant := time.Now()
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	for {
		select {
		case agg := <-write_chan:

			cap := final_agg_size(agg)

			if cap > int(batchingmaxsize) {
				for i := range agg {
					// FIXME uncomment
					ack_chan <- agg[i]
				}
				n_ignored++
				// TODO metric ignored
			} else {
				producer.SendAsync(
					context.Background(),
					&pulsar.ProducerMessage{
						Payload: json_write_messages(agg, cap),
						Key:     agg[0].Key(),
					},
					write_callback(agg, ack_chan),
				)
				n_writen += 1
			}
		case <-tick.C:
			since := time.Since(last_instant)
			last_instant = time.Now()
			logrus.Infof("Write rate: %.3f msg/s (ignored: %d msg/min)", n_writen/float64(since/time.Second), n_ignored)
			n_writen = 0
			n_ignored = 0
		}
	}
}

/*
 * Helper functions
 */

func final_agg_size(messages []pulsar.ConsumerMessage) int {
	size := len(messages)*2 + 100 // { "transaction": [..., ..., ..., ...] }
	for i := 0; i < len(messages); i++ {
		size += len((messages)[i].Payload())
	}
	return size
}

func json_write_messages(messages []pulsar.ConsumerMessage, max_cap int) []byte {
	out := make([]byte, 0, max_cap)
	out = append(out, "["...)
	for i := 0; i < len(messages); i++ {
		out = append(out, (messages)[i].Payload()...)
		if i+1 < len(messages) {
			out = append(out, ", "...)
		}
	}
	out = append(out, "]"...)
	return out
}
