package d2lkinesis

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_D2lKinesisOutput_PutRecords_SingleRecord_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(1, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
	})
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecords_SingleRecord_Failure(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(0, 1)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
	})
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecords_SingleRecord_Error(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupErrorResponse(fmt.Errorf("timeout"))

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
	})

	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecords_MultipleRecords_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(1, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
		record2,
	})
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecords_MultipleRecords_PartialSuccess(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(2, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultErrorEntry(
			"ProvisionedThroughputExceededException",
			"Rate exceeded for shard shardId-000000000001 in stream exampleStreamName under account 111111111111.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000003",
			"49543463076570308322303623326179887152428262250726293522",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
	})

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})
	record3 := createTestKinesisRecord(1, []byte{0x02})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
		record2,
		record3,
	})
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
		record3,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
				record3.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecords_MultipleRecords_Failure(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(0, 2)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	failedRecords := output.putRecords([]*kinesisRecord{
		record1,
		record2,
	})
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
		record2,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
	})
}

// ---------------------------------------------------------------------------------

func Test_D2lKinesisOutput_PutRecordBatches_SingleBatch_SingleRecord_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(1, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
	))
	assert.NoError(err, "Should not error")
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_SingleBatch_SingleRecord_Failure(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(0, 1)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
	))
	assert.NoError(err, "Should not error")
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_SingleBatch_MultipleRecords_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(1, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
		record2,
	))
	assert.NoError(err, "Should not error")
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_SingleBatch_MultipleRecords_PartialSucess(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultEntry(
			"shardId-000000000000",
			"49543463076548007577105092703039560359975228518395012686",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000003",
			"49543463076570308322303623326179887152428262250726293522",
		),
	})

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})
	record3 := createTestKinesisRecord(1, []byte{0x03})

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
		record2,
		record3,
	))
	assert.NoError(err, "Should not error")
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record2,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
				record3.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_SingleBatch_MultipleRecords_Failure(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(0, 2)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
		record2,
	))
	assert.NoError(err, "Should not error")
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
		record2,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_MultipleBatches_RequestSizeLimit_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(4, 0)
	svc.SetupGenericResponse(2, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	// batch 1
	record1 := createTestKinesisRecord(100, make([]byte, awsMaxRecordSize))
	record2 := createTestKinesisRecord(200, make([]byte, awsMaxRecordSize))
	record3 := createTestKinesisRecord(300, make([]byte, awsMaxRecordSize))
	record4 := createTestKinesisRecord(400, make([]byte, awsMaxRecordSize))
	// batch 2
	record5 := createTestKinesisRecord(500, make([]byte, awsMaxRecordSize))
	record6 := createTestKinesisRecord(600, make([]byte, awsMaxRecordSize))

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
		record2,
		record3,
		record4,
		record5,
		record6,
	))
	assert.NoError(err, "Should not error")
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
				record3.Entry,
				record4.Entry,
			},
		},
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record5.Entry,
				record6.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_MultipleBatches_RequestSizeLimit_PartialSuccess(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(2, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultErrorEntry(
			"ProvisionedThroughputExceededException",
			"Rate exceeded for shard shardId-000000000001 in stream exampleStreamName under account 111111111111.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000003",
			"49543463076570308322303623326179887152428262250726293522",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000000",
			"49543463076548007577105092703039560359975228518395012686",
		),
	})
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultErrorEntry(
			"ProvisionedThroughputExceededException",
			"Rate exceeded for shard shardId-000000000002 in stream exampleStreamName under account 111111111111.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000001",
			"49543463076570308322303623326179887152428262250726293522",
		),
	})

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	// batch 1
	record1 := createTestKinesisRecord(100, make([]byte, awsMaxRecordSize))
	record2 := createTestKinesisRecord(200, make([]byte, awsMaxRecordSize))
	record3 := createTestKinesisRecord(300, make([]byte, awsMaxRecordSize))
	record4 := createTestKinesisRecord(400, make([]byte, awsMaxRecordSize))
	// batch 2
	record5 := createTestKinesisRecord(500, make([]byte, awsMaxRecordSize))
	record6 := createTestKinesisRecord(600, make([]byte, awsMaxRecordSize))

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(
		record1,
		record2,
		record3,
		record4,
		record5,
		record6,
	))
	assert.NoError(err, "Should not error")
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		record1,
		record3,
		record5,
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
				record3.Entry,
				record4.Entry,
			},
		},
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record5.Entry,
				record6.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_MultipleBatches_RecordsLimit_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(awsMaxRecordsPerRequest, 0)
	svc.SetupGenericResponse(1, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	records := createTestKinesisRecords(
		[]byte{0x01},
		awsMaxRecordsPerRequest+1,
	)

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(records...))
	assert.NoError(err, "Should not error")
	assert.Empty(failedRecords, "Should not return any failed records")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: selectPutRecordsRequestEntries(
				records[0:awsMaxRecordsPerRequest],
			),
		},
		{
			StreamName: &streamName,
			Records: selectPutRecordsRequestEntries(
				records[awsMaxRecordsPerRequest:],
			),
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatches_MultipleBatches_RecordsLimit_PartialSuccess(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(awsMaxRecordsPerRequest-1, 1)
	svc.SetupGenericResponse(0, 1)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	records := createTestKinesisRecords(
		[]byte{0x01},
		awsMaxRecordsPerRequest+1,
	)

	failedRecords, err := output.putRecordBatches(createKinesisRecordSet(records...))
	assert.NoError(err, "Should not error")
	assertFailedRecords(assert, failedRecords, []*kinesisRecord{
		records[awsMaxRecordsPerRequest-1],
		records[awsMaxRecordsPerRequest],
	})

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: selectPutRecordsRequestEntries(
				records[0:awsMaxRecordsPerRequest],
			),
		},
		{
			StreamName: &streamName,
			Records: selectPutRecordsRequestEntries(
				records[awsMaxRecordsPerRequest:],
			),
		},
	})
}

// ---------------------------------------------------------------------------------

func Test_D2lKinesisOutput_PutRecordBatchesWithRetry_Success(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(2, 0)

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	err := output.putRecordBatchesWithRetry(createKinesisRecordSet(
		record1,
		record2,
	))
	assert.NoError(err, "Should not error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatchesWithRetry_SingleRetry(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultEntry(
			"shardId-000000000000",
			"49543463076548007577105092703039560359975228518395012686",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
	})
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultEntry(
			"shardId-000000000001",
			"49543463076570308322303623326179887152428262250726293522",
		),
	})

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 4)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})

	err := output.putRecordBatchesWithRetry(createKinesisRecordSet(
		record1,
		record2,
	))
	assert.NoError(err, "Should not error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
			},
		},
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record2.Entry,
			},
		},
	})
}

func Test_D2lKinesisOutput_PutRecordBatchesWithRetry_MaxRetries(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(2, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
		createPutRecordsResultEntry(
			"shardId-000000000000",
			"49543463076548007577105092703039560359975228518395012686",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
	})
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultEntry(
			"shardId-000000000001",
			"49543463076570308322303623326179887152428262250726293522",
		),
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
	})
	svc.SetupResponse(1, []*kinesis.PutRecordsResultEntry{
		createPutRecordsResultErrorEntry(
			"InternalFailure",
			"Internal service failure.",
		),
	})

	streamName := "stream"
	output := createTestKinesisOutput(svc, streamName, 2)

	record1 := createTestKinesisRecord(1, []byte{0x01})
	record2 := createTestKinesisRecord(1, []byte{0x02})
	record3 := createTestKinesisRecord(1, []byte{0x03})

	err := output.putRecordBatchesWithRetry(createKinesisRecordSet(
		record1,
		record2,
		record3,
	))
	assert.NoError(err, "Should not error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record2.Entry,
				record3.Entry,
			},
		},
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record1.Entry,
				record3.Entry,
			},
		},
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				record3.Entry,
			},
		},
	})
}

// ---------------------------------------------------------------------------------

func Test_D2lKinesisOutput_Write_NoMetrics(t *testing.T) {
	assert := assert.New(t)

	output := &d2lKinesisOutput{}
	err := output.Write(nil)
	assert.NoError(err, "Should not error")
}

func Test_D2lKinesisOutput_Write_MultipleMetrics(t *testing.T) {
	assert := assert.New(t)

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(2, 0)

	streamName := "stream"
	partitionKey := testPartitionKey

	output := createTestKinesisOutput(svc, streamName, 4)
	output.recordGenerator, _ = createGZipKinesisRecordGenerator(
		testutil.Logger{},
		awsMaxRecordSize,
		testPartitionKeyProvider,
		influxSerializer,
	)

	metric1, metric1Data := createTestMetric(t, "metric1", influxSerializer)
	metric2, metric2Data := createTestMetric(t, "metric2", influxSerializer)

	err := output.Write([]telegraf.Metric{
		metric1,
		metric2,
	})
	assert.NoError(err, "Should not error")

	data, dataErr := gzipCompressBlocks([][]byte{
		concatByteSlices(metric1Data, metric2Data),
	})
	assert.NoError(dataErr, "Should compress data")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				{
					PartitionKey: &partitionKey,
					Data:         data,
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------------

func Test_D2lKinesisOutput_Connect_MaxRecordSizeLessThan1000Bytes(t *testing.T) {
	assert := assert.New(t)

	output := d2lKinesisOutput{
		MaxRecordSize: 999,
		StreamName:    "metrics",
	}

	err := output.Connect()
	assert.Equal(err.Error(), "max_record_size must be at least 1000 bytes")
}

func Test_D2lKinesisOutput_Connect_MaxRecordSizeGreaterThan1MiB(t *testing.T) {
	assert := assert.New(t)

	output := d2lKinesisOutput{
		MaxRecordSize: awsMaxRecordSize + 1,
		StreamName:    "metrics",
	}

	err := output.Connect()
	assert.Equal(err.Error(), "max_record_size must be less than or equal to the aws limit of 1048576 bytes")
}

func Test_D2lKinesisOutput_Connect_EmptyStreamName(t *testing.T) {
	assert := assert.New(t)

	output := d2lKinesisOutput{
		MaxRecordSize: awsMaxRecordSize,
		StreamName:    "",
	}

	err := output.Connect()
	assert.Equal(err.Error(), "stream_name is required")
}

// ---------------------------------------------------------------------------------

func assertFailedRecords(
	assert *assert.Assertions,
	actual []*kinesisRecord,
	expected []*kinesisRecord,
) {

	if !assert.Equal(
		len(expected),
		len(actual),
		fmt.Sprintf("Expected %v failed records", len(expected)),
	) {
		return
	}

	for i, expectedRecord := range expected {
		assert.Same(
			expectedRecord,
			actual[i],
			i,
			fmt.Sprintf("Failed record at index %d should be as expected", i),
		)
	}
}

func createPutRecordsResultEntry(
	shardID string,
	sequenceNumber string,
) *kinesis.PutRecordsResultEntry {

	return &kinesis.PutRecordsResultEntry{
		ShardId:        &shardID,
		SequenceNumber: &sequenceNumber,
	}
}

func createPutRecordsResultErrorEntry(
	errorCode string,
	errorMessage string,
) *kinesis.PutRecordsResultEntry {

	return &kinesis.PutRecordsResultEntry{
		ErrorCode:    &errorCode,
		ErrorMessage: &errorMessage,
	}
}

func createTestKinesisOutput(
	svc *mockKinesisPutRecords,
	streamName string,
	maxRecordRetries uint8,
) d2lKinesisOutput {

	return d2lKinesisOutput{
		Log:              testutil.Logger{},
		StreamName:       streamName,
		svc:              svc,
		MaxRecordRetries: maxRecordRetries,
	}
}

func selectPutRecordsRequestEntries(
	records []*kinesisRecord,
) []*kinesis.PutRecordsRequestEntry {

	count := len(records)
	entries := make([]*kinesis.PutRecordsRequestEntry, count)

	for i, record := range records {
		entries[i] = record.Entry
	}

	return entries
}

// ---------------------------------------------------------------------------------

type mockKinesisPutRecordsResponse struct {
	Output *kinesis.PutRecordsOutput
	Err    error
}

type mockKinesisPutRecords struct {
	kinesisiface.KinesisAPI

	requests  []*kinesis.PutRecordsInput
	responses []*mockKinesisPutRecordsResponse
}

func (m *mockKinesisPutRecords) SetupResponse(
	failedRecordCount int64,
	records []*kinesis.PutRecordsResultEntry,
) {

	m.responses = append(m.responses, &mockKinesisPutRecordsResponse{
		Err: nil,
		Output: &kinesis.PutRecordsOutput{
			FailedRecordCount: &failedRecordCount,
			Records:           records,
		},
	})
}

func (m *mockKinesisPutRecords) SetupGenericResponse(
	successfulRecordCount uint32,
	failedRecordCount uint32,
) {

	errorCode := "InternalFailure"
	errorMessage := "Internal Service Failure"
	shard := "shardId-000000000003"

	records := []*kinesis.PutRecordsResultEntry{}

	for i := uint32(0); i < successfulRecordCount; i++ {
		sequenceNumber := fmt.Sprintf("%d", i)
		records = append(records, &kinesis.PutRecordsResultEntry{
			SequenceNumber: &sequenceNumber,
			ShardId:        &shard,
		})
	}

	for i := uint32(0); i < failedRecordCount; i++ {
		records = append(records, &kinesis.PutRecordsResultEntry{
			ErrorCode:    &errorCode,
			ErrorMessage: &errorMessage,
		})
	}

	m.SetupResponse(int64(failedRecordCount), records)
}

func (m *mockKinesisPutRecords) SetupErrorResponse(err error) {

	m.responses = append(m.responses, &mockKinesisPutRecordsResponse{
		Err:    err,
		Output: nil,
	})
}

func (m *mockKinesisPutRecords) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {

	reqNum := len(m.requests)
	if reqNum > len(m.responses) {
		return nil, fmt.Errorf("Response for request %+v not setup", reqNum)
	}

	m.requests = append(m.requests, input)

	resp := m.responses[reqNum]
	return resp.Output, resp.Err
}

func (m *mockKinesisPutRecords) AssertRequests(
	assert *assert.Assertions,
	expected []*kinesis.PutRecordsInput,
) {

	if !assert.Equal(
		len(expected),
		len(m.requests),
		fmt.Sprintf("Expected %v requests", len(expected)),
	) {
		return
	}

	for i, expectedInput := range expected {
		actualInput := m.requests[i]

		assert.Equal(
			expectedInput.StreamName,
			actualInput.StreamName,
			fmt.Sprintf("Expected request %v to have correct StreamName", i),
		)

		assert.Equal(
			len(expectedInput.Records),
			len(actualInput.Records),
			fmt.Sprintf("Expected request %v to have %v Records", i, len(expectedInput.Records)),
		)

		for r, expectedRecord := range expectedInput.Records {
			actualRecord := actualInput.Records[r]

			assert.Equal(
				&expectedRecord.PartitionKey,
				&actualRecord.PartitionKey,
				fmt.Sprintf("Expected (request %v, record %v) to have correct PartitionKey", i, r),
			)

			assert.Equal(
				&expectedRecord.ExplicitHashKey,
				&actualRecord.ExplicitHashKey,
				fmt.Sprintf("Expected (request %v, record %v) to have correct ExplicitHashKey", i, r),
			)

			assert.Equal(
				expectedRecord.Data,
				actualRecord.Data,
				fmt.Sprintf("Expected (request %v, record %v) to have correct Data", i, r),
			)
		}
	}
}
