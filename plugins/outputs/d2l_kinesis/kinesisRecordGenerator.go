package d2lkinesis

import "github.com/influxdata/telegraf"

type kinesisRecordGenerator interface {
	kinesisRecordIterator

	Reset(metrics []telegraf.Metric)
}
