package d2lkinesis

import (
	"bytes"
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

	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metricData),
		record1,
	)
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

	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(
			2,
			concatByteSlices(metric1Data, metric2Data),
		),
		record1,
	)
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

	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metric1Data),
		record1,
	)
	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metric2Data),
		record2,
	)
}

func Test_GZipKinesisRecordGenerator_AtMax(t *testing.T) {
	assert := assert.New(t)

	metric1 := testutil.TestMetric(1, "metric1")
	metric1Data := []byte{0xa1, 0xb2, 0xc3, 0xd4}

	metric2 := testutil.TestMetric(1, "metric2")
	metric2Data := []byte{0x50, 0x71, 0xe5, 0xf6}

	metric3 := testutil.TestMetric(1, "metric3")
	metric3Data := []byte{0x70}

	mockSerializer := createMockMetricSerializer()
	mockSerializer.SetupMetricData(metric1, metric1Data)
	mockSerializer.SetupMetricData(metric2, metric2Data)
	mockSerializer.SetupMetricData(metric3, metric3Data)

	const maxRecordSize int = 34
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
	assert.NotNil(record2, "Should read second record")

	record3, err := generator.Next()
	assert.NoError(err, "Next should not error")
	assert.NotNil(record2, "Should read third record")

	assertEndOfIterator(assert, generator)

	assert.Equal(
		maxRecordSize,
		len(record1.Entry.Data),
		"Should serialize first metric to max record size",
	)
	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metric1Data),
		record1,
	)

	assert.Less(
		len(record2.Entry.Data),
		maxRecordSize,
		"Should serialize second metric to less than max record size",
	)
	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metric2Data),
		record2,
	)

	assert.Less(
		len(record3.Entry.Data),
		maxRecordSize,
		"Should serialize third metric to less than max record size",
	)
	assertGZippedKinesisRecord(
		assert,
		createTestKinesisRecord(1, metric3Data),
		record3,
	)
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
	expected *kinesisRecord,
	actual *kinesisRecord,
) {

	if actual == nil {
		assert.NotNil(actual, "Actual kinesis record should not be nil")
		return
	}

	actualDecompressedData, decompressErr := gzipDecompress(
		actual.Entry.Data,
		len(expected.Entry.Data),
	)
	if decompressErr != nil {
		assert.NoError(decompressErr, "Actual Entry.Data should have decompressed")
		return
	}

	assert.Equal(
		base64.StdEncoding.EncodeToString(expected.Entry.Data),
		base64.StdEncoding.EncodeToString(actualDecompressedData),
		"Entry.Data should be as expected when decompressed",
	)

	assert.Nil(
		expected.Entry.ExplicitHashKey,
		"Entry.ExplicitHashKey should not be expected",
	)
	assert.Nil(
		expected.Entry.ExplicitHashKey,
		"Entry.ExplicitHashKey should not be assigned",
	)

	assert.Equal(
		*expected.Entry.PartitionKey,
		*actual.Entry.PartitionKey,
		"Entry.PartitionKey should be as expected",
	)

	assert.Equal(
		expected.Metrics,
		actual.Metrics,
		"Metrics should be as expected",
	)
}

func gzipDecompress(
	data []byte,
	bufferSize int,
) ([]byte, error) {

	compressedReader := bytes.NewReader(data)

	gzipReader, gzipReaderErr := gzip.NewReader(compressedReader)
	if gzipReaderErr != nil {
		return nil, gzipReaderErr
	}

	result := make([]byte, 0, bufferSize)
	buffer := make([]byte, bufferSize)

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
