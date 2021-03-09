package d2lkinesis

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"math"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const gzipHeaderSize = 10
const gzipFooterSize = 8

func createGZipKinesisRecordGenerator(
	log telegraf.Logger,
	maxRecordSize int,
	pkGenerator partitionKeyGenerator,
	serializer serializers.Serializer,
) (kinesisRecordGenerator, error) {

	buffer := bytes.NewBuffer([]byte{})

	writer, writerErr := gzip.NewWriterLevel(buffer, flate.BestCompression)
	if writerErr != nil {
		return nil, writerErr
	}

	generator := &gzipKinesisRecordGenerator{
		log:           log,
		maxRecordSize: maxRecordSize,
		pkGenerator:   pkGenerator,
		serializer:    serializer,

		buffer: buffer,
		writer: writer,
	}

	return generator, nil
}

type gzipKinesisRecordGenerator struct {
	kinesisRecordIterator

	buffer        *bytes.Buffer
	log           telegraf.Logger
	maxRecordSize int
	pkGenerator   partitionKeyGenerator
	serializer    serializers.Serializer
	writer        *gzip.Writer

	index        int
	metricsCount int
	metrics      []telegraf.Metric
}

func (g *gzipKinesisRecordGenerator) Reset(
	metrics []telegraf.Metric,
) {

	g.index = 0
	g.metrics = metrics
	g.metricsCount = len(metrics)
}

func (g *gzipKinesisRecordGenerator) yieldRecord(
	metrics int,
) (*kinesisRecord, error) {

	closeErr := g.writer.Close()
	if closeErr != nil {
		return nil, closeErr
	}

	bufferBytes := g.buffer.Bytes()
	data := make([]byte, len(bufferBytes))
	copy(data, bufferBytes)

	partitionKey := g.pkGenerator()

	entry := &kinesis.PutRecordsRequestEntry{
		Data:         data,
		PartitionKey: &partitionKey,
	}

	record := createKinesisRecord(entry, metrics)

	return record, nil
}

func (g *gzipKinesisRecordGenerator) Next() (*kinesisRecord, error) {

	startIndex := g.index
	if startIndex >= g.metricsCount {
		return nil, nil
	}

	index := startIndex
	recordMetricCount := 0
	recordSizeEstimator := createGZipSizeEstimator()

	g.buffer.Reset()
	g.writer.Reset(g.buffer)

	for ; index < g.metricsCount; index++ {
		metric := g.metrics[index]

		bytes, serializeErr := g.serializer.Serialize(metric)
		if serializeErr != nil {

			g.log.Errorf(
				"Failed to serialize metric: %s",
				serializeErr.Error(),
			)
			continue
		}

		bytesCount := len(bytes)

		if recordSizeEstimator.MaxPotentialSizeWith(bytesCount) > g.maxRecordSize {

			flushErr := g.writer.Flush()
			if flushErr != nil {
				return nil, flushErr
			}

			// commit the flushed buffer length
			recordSizeEstimator.Commit(g.buffer.Len())

			if recordSizeEstimator.MaxPotentialSizeWith(bytesCount) > g.maxRecordSize {

				if recordMetricCount == 0 {
					g.log.Warnf(
						"Dropping excessively large '%s' metric",
						metric.Name(),
					)
					continue
				}

				g.index = index
				return g.yieldRecord(recordMetricCount)
			}
		}

		count, writeErr := g.writer.Write(bytes)
		if writeErr != nil {
			return nil, writeErr
		}

		recordMetricCount++
		recordSizeEstimator.RecordBytes(count)
	}

	if recordMetricCount > 0 {
		g.index = index + 1
		return g.yieldRecord(recordMetricCount)
	}

	return nil, nil
}

func createGZipSizeEstimator() gZipSizeEstimator {
	return gZipSizeEstimator{
		bytesCommited: gzipHeaderSize + gzipFooterSize,
		bytesPending:  0,
	}
}

type gZipSizeEstimator struct {
	bytesCommited int
	bytesPending  int
}

func (e *gZipSizeEstimator) RecordBytes(bytes int) {
	e.bytesPending += bytes
}

func (e *gZipSizeEstimator) Commit(bytes int) {
	e.bytesCommited = bytes + gzipFooterSize
	e.bytesPending = 0
}

func (e *gZipSizeEstimator) MaxPotentialSizeWith(additionalBytes int) int {

	bytesCount := e.bytesPending + additionalBytes
	blocksOverhead := 5 * int((math.Floor(float64(bytesCount)/16383) + 1))

	return e.bytesCommited + bytesCount + blocksOverhead
}
