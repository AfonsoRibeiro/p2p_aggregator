package aggregator

import (
	"p2p_aggregator/src/prom_metrics"

	"github.com/apache/pulsar-client-go/pulsar"
)

type container struct {
	min_publish_time int64
	max_publish_time int64
	agg_size         uint64
	messages         []pulsar.ConsumerMessage
}

type mapFunctions struct {
	time_delta     int64
	total_size     uint64
	max_total_size uint64
	write_chan     chan<- []pulsar.ConsumerMessage
}

func (mf *mapFunctions) ExpiredHandler(key *string, value *container) {
	mf.total_size -= value.agg_size
	mf.write_chan <- value.messages

	// FIXME should be in ack but ack not done in end to to reconnect (fixme)
	prom_metrics.Prom_metric.Sub_number_inflight_msg(len(value.messages))
}

func (mf *mapFunctions) CapacityRule(curcapacity uint64, curlen uint64, head *container, tail *container) uint64 {
	delta := tail.max_publish_time - head.min_publish_time

	if delta > mf.time_delta { // newest <-> oldest
		curcapacity = curlen - 1
	} else if mf.total_size > mf.max_total_size {
		curcapacity = curlen - 1
	} else if curcapacity > 0 {
		curcapacity = 0
	}

	return curcapacity
}
