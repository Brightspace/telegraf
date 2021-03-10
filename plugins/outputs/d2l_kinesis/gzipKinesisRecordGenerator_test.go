package d2lkinesis

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var influxSerializer serializers.Serializer = influx.NewSerializer()

func Test_CreateGZipKinesisRecordGenerator(t *testing.T) {
	assert := assert.New(t)

	generator, err := createGZipKinesisRecordGenerator(
		testutil.Logger{},
		256,
		testPartitionKeyProvider,
		influxSerializer,
	)

	assert.NoError(err)
	assert.NotNil(generator)
}

func Test_GZipKinesisRecordGenerator_ZeroRecords(t *testing.T) {
	assert := assert.New(t)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, influxSerializer)
	generator.Reset([]telegraf.Metric{})

	assertEndOfIterator(assert, generator)
}

func Test_GZipKinesisRecordGenerator_SingleMetric_SingleRecord(t *testing.T) {
	assert := assert.New(t)

	metric, metricData := createTestMetric(t, "test", influxSerializer)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, influxSerializer)
	generator.Reset([]telegraf.Metric{metric})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1)

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metricData,
	})
}

func Test_GZipKinesisRecordGenerator_TwoMetrics_SingleRecord(t *testing.T) {
	assert := assert.New(t)

	metric1, metric1Data := createTestMetric(t, "metric1", influxSerializer)
	metric2, metric2Data := createTestMetric(t, "metric2", influxSerializer)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, influxSerializer)
	generator.Reset([]telegraf.Metric{metric1, metric2})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1)

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		concatByteSlices(metric1Data, metric2Data),
	})
}

func Test_GZipKinesisRecordGenerator_TwoMetrics_TwoRecords(t *testing.T) {
	assert := assert.New(t)

	metric1, metric1Data := createTestMetric(t, "metric1", influxSerializer)
	metric2, metric2Data := createTestMetric(t, "metric2", influxSerializer)

	generator := createTestGZipKinesisRecordGenerator(t, 92, influxSerializer)
	generator.Reset([]telegraf.Metric{metric1, metric2})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1)

	record2, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record2)

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metric1Data,
		{}, // empty block after flush
	})

	assertGZippedKinesisRecord(assert, record2, [][]byte{
		metric2Data,
	})
}

func Test_GZipKinesisRecordGenerator_UnderMaxWithoutFlush(t *testing.T) {
	assert := assert.New(t)

	metric := testutil.TestMetric(1, "metric")
	metricData := []byte{0xa1, 0xb2, 0xc3, 0xd4, 0xe5}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric, metricData)

	const maxRecordSize int = 35
	generator := createTestGZipKinesisRecordGenerator(t, maxRecordSize, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric,
	})

	record, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record, "Should read first record")

	assertEndOfIterator(assert, generator)

	assert.Equal(
		maxRecordSize-gzipFlushBlockSize,
		len(record.Entry.Data),
		"Should serialize first metric to ( max record size - flush block size )",
	)
	assertGZippedKinesisRecord(assert, record, [][]byte{
		metricData,
	})
}

func Test_GZipKinesisRecordGenerator_UnderMaxWithFlush(t *testing.T) {
	assert := assert.New(t)

	metric1 := testutil.TestMetric(1, "metric1")
	metric1Data := []byte{0xa1, 0xb2, 0xc3, 0xd4}

	metric2 := testutil.TestMetric(1, "metric2")
	metric2Data := []byte{0xf6, 0xe5, 0xd4, 0xc3, 0xb2}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricData(metric2, metric2Data)

	const maxRecordSize int = 35
	generator := createTestGZipKinesisRecordGenerator(t, maxRecordSize, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric1,
		metric2,
	})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read first record")

	record2, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record2, "Should read second record")

	assertEndOfIterator(assert, generator)

	assert.Equal(
		maxRecordSize-1,
		len(record1.Entry.Data),
		"Should serialize first metric to ( max record size - 1 )",
	)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metric1Data,
		{}, // empty block after flush
	})

	assertGZippedKinesisRecord(assert, record2, [][]byte{
		metric2Data,
	})
}

func Test_GZipKinesisRecordGenerator_AtMaxWithFlush(t *testing.T) {
	assert := assert.New(t)

	metric1 := testutil.TestMetric(1, "metric1")
	metric1Data := []byte{0xa1, 0xb2, 0xc3, 0xd4, 0xe5}

	metric2 := testutil.TestMetric(1, "metric2")
	metric2Data := []byte{0xf6, 0xe5, 0xd4, 0xc3, 0xb2}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricData(metric2, metric2Data)

	const maxRecordSize int = 35
	generator := createTestGZipKinesisRecordGenerator(t, maxRecordSize, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric1,
		metric2,
	})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read first record")

	record2, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record2, "Should read second record")

	assertEndOfIterator(assert, generator)

	assert.Equal(
		maxRecordSize,
		len(record1.Entry.Data),
		"Should serialize first metric to max record size",
	)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metric1Data,
		{}, // empty block after flush
	})

	assertGZippedKinesisRecord(assert, record2, [][]byte{
		metric2Data,
	})
}

func Test_GZipKinesisRecordGenerator_OverMaxWithFlush_FirstMetric(t *testing.T) {
	assert := assert.New(t)

	metric1 := testutil.TestMetric(1, "metric1")
	metric1Data := []byte{0xa1, 0xb2, 0xc3, 0xd4, 0xe5, 0xf6}

	metric2 := testutil.TestMetric(1, "metric2")
	metric2Data := []byte{0xf6, 0xe5, 0xd4, 0xc3, 0xb2}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricData(metric2, metric2Data)

	const maxRecordSize int = 35
	generator := createTestGZipKinesisRecordGenerator(t, maxRecordSize, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric1,
		metric2,
	})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read first record")

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metric2Data,
	})
}

func Test_GZipKinesisRecordGenerator_OverMaxWithFlush_SecondMetric(t *testing.T) {
	assert := assert.New(t)

	metric1 := testutil.TestMetric(1, "metric1")
	metric1Data := []byte{0x01, 0x02}

	metric2 := testutil.TestMetric(2, "metric2")
	metric2Data := []byte{0xa1, 0xb2, 0xc3, 0xd4, 0xe5, 0xf6}

	metric3 := testutil.TestMetric(3, "metric3")
	metric3Data := []byte{0xf6, 0xe5, 0xd4, 0xc3, 0xb2}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricData(metric2, metric2Data)
	mockSerializer.SetupMetricData(metric3, metric3Data)

	const maxRecordSize int = 35
	generator := createTestGZipKinesisRecordGenerator(t, maxRecordSize, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric1,
		metric2,
		metric3,
	})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read first record")

	record2, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read second record")

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		metric1Data,
		{}, // empty block after flush
	})

	assertGZippedKinesisRecord(assert, record2, [][]byte{
		metric3Data,
	})
}

func Test_GZipKinesisRecordGenerator_SingleMetric_ValidGZip(t *testing.T) {
	assert := assert.New(t)

	metric, metricData := createTestMetric(t, "test", influxSerializer)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, influxSerializer)
	generator.Reset([]telegraf.Metric{metric})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1)

	assertEndOfIterator(assert, generator)

	decompressed, decompressErr := gzipDecompress(record1.Entry.Data)
	assert.NoError(decompressErr, "Should decompress data")
	assert.Equal(
		base64.StdEncoding.EncodeToString(metricData),
		base64.StdEncoding.EncodeToString(decompressed),
		"Decompressed data should equal original metric bytes",
	)
}

func Test_GZipKinesisRecordGenerator_MultipleMetrics_ValidGZip(t *testing.T) {
	assert := assert.New(t)

	metric1, metric1Data := createTestMetric(t, "metric1", influxSerializer)
	metric2, metric2Data := createTestMetric(t, "metric2", influxSerializer)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, influxSerializer)
	generator.Reset([]telegraf.Metric{metric1, metric2})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1)

	assertEndOfIterator(assert, generator)

	decompressed, decompressErr := gzipDecompress(record1.Entry.Data)
	assert.NoError(decompressErr, "Should decompress data")
	assert.Equal(
		base64.StdEncoding.EncodeToString(concatByteSlices(metric1Data, metric2Data)),
		base64.StdEncoding.EncodeToString(decompressed),
		"Decompressed data should equal original metric bytes",
	)
}

func Test_GZipKinesisRecordGenerator_SerializerError(t *testing.T) {
	assert := assert.New(t)

	metric1, metric1Data := createTestMetric(t, "metric1", influxSerializer)
	metric2 := testutil.TestMetric(1, "metric2")
	metric3, metric3Data := createTestMetric(t, "metric3", influxSerializer)

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricError(metric2, fmt.Errorf("boom"))
	mockSerializer.SetupMetricData(metric3, metric3Data)

	generator := createTestGZipKinesisRecordGenerator(t, 1024, &mockSerializer)
	generator.Reset([]telegraf.Metric{
		metric1,
		metric2,
		metric3,
	})

	record1, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record1, "Should read first record")

	assertEndOfIterator(assert, generator)

	assertGZippedKinesisRecord(assert, record1, [][]byte{
		concatByteSlices(metric1Data, metric3Data),
	})
}

// ---------------------------------------------------------------------------------

func createTestGZipKinesisRecordGenerator(
	t *testing.T,
	maxRecordSize int,
	serializer serializers.Serializer,
) kinesisRecordGenerator {

	generator, err := createGZipKinesisRecordGenerator(
		testutil.Logger{},
		maxRecordSize,
		testPartitionKeyProvider,
		serializer,
	)
	require.NoError(t, err)

	return generator
}

func createTestMetric(
	t *testing.T,
	name string,
	serializer serializers.Serializer,
) (telegraf.Metric, []byte) {

	metric := testutil.TestMetric(1, name)

	data, err := serializer.Serialize(metric)
	require.NoError(t, err)

	return metric, data
}

func assertGZippedKinesisRecord(
	assert *assert.Assertions,
	actual *kinesisRecord,
	expectedGZipBlocks [][]byte,
) {

	if actual == nil {
		assert.NotNil(actual, "Kinesis record should not be nil")
		return
	}

	entry := actual.Entry
	if entry == nil {
		assert.NotNil(actual, "Entry should not be nil")
		return
	}

	assert.Nil(
		entry.ExplicitHashKey,
		"Entry.ExplicitHashKey should not be expected",
	)

	partitionKey := entry.PartitionKey
	if partitionKey == nil {
		assert.NotNil(actual, "Entry.PartitionKey should not be nil")
		return
	}
	assert.Equal(*partitionKey, testPartitionKey, "Entry.PartitionKey should be as expected")

	entryData := entry.Data
	if entry.Data == nil {
		assert.NotNil(entryData, "Entry.Data should not be nil")
		return
	}

	expectedData, compressErr := gzipCompressBlocks(expectedGZipBlocks)
	if compressErr != nil {
		assert.NoError(compressErr, "Should be able to compress expected blocks")
		return
	}

	assert.Equal(
		base64.StdEncoding.EncodeToString(expectedData),
		base64.StdEncoding.EncodeToString(entryData),
		"Entry.Data should be as expected",
	)
}

func gzipCompressBlocks(blocks [][]byte) ([]byte, error) {

	buffer := bytes.NewBuffer([]byte{})

	writer, writerErr := gzip.NewWriterLevel(buffer, flate.BestCompression)
	if writerErr != nil {
		return nil, writerErr
	}

	for index, block := range blocks {

		if index > 0 {

			flushErr := writer.Flush()
			if flushErr != nil {
				return nil, flushErr
			}
		}

		_, writeErr := writer.Write(block)
		if writeErr != nil {
			return nil, writeErr
		}
	}

	closeErr := writer.Close()
	if closeErr != nil {
		return nil, closeErr
	}

	return buffer.Bytes(), nil
}

func gzipDecompress(data []byte) ([]byte, error) {

	compressedReader := bytes.NewReader(data)

	gzipReader, gzipReaderErr := gzip.NewReader(compressedReader)
	if gzipReaderErr != nil {
		return nil, gzipReaderErr
	}

	result := make([]byte, 0, 128)
	buffer := make([]byte, 128)

	for {
		readCount, readErr := gzipReader.Read(buffer)

		if readCount > 0 {
			result = append(result, buffer[0:readCount]...)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, readErr
		}
	}

	closeErr := gzipReader.Close()
	if closeErr != nil {
		return nil, closeErr
	}

	return result, nil
}

func concatByteSlices(slices ...[]byte) []byte {

	size := 0
	for i := 0; i < len(slices); i++ {
		size += len(slices[i])
	}

	result := make([]byte, 0, size)
	for i := 0; i < len(slices); i++ {
		result = append(result, slices[i]...)
	}

	return result
}

func createMockMetricSerializer() mockMetricSerializer {
	return mockMetricSerializer{
		results: make(map[telegraf.Metric]mockMetricSerializerResult),
	}
}

type mockMetricSerializer struct {
	serializers.Serializer

	results map[telegraf.Metric]mockMetricSerializerResult
}

type mockMetricSerializerResult struct {
	data []byte
	err  error
}

func (s *mockMetricSerializer) SetupMetricData(metric telegraf.Metric, data []byte) {
	s.results[metric] = mockMetricSerializerResult{data: data}
}

func (s *mockMetricSerializer) SetupMetricError(metric telegraf.Metric, err error) {
	s.results[metric] = mockMetricSerializerResult{err: err}
}

func (s *mockMetricSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	result := s.results[metric]

	if result.err != nil {
		return nil, result.err
	}

	if result.data == nil {
		return nil, fmt.Errorf("Metric '%s' serialization not mocked", metric.Name())
	}

	return result.data, nil
}

func (s *mockMetricSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	batch := []byte{}

	for _, metric := range metrics {

		result := s.results[metric]

		if result.err != nil {
			return nil, result.err
		}

		if result.data == nil {
			return nil, fmt.Errorf("Metric '%s' serialization not mocked", metric.Name())
		}

		batch = append(batch, result.data...)
	}

	return batch, nil
}
