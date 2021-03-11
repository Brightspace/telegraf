package d2lkinesis

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/assert"
)

const testPartitionKey string = "abc"

func testPartitionKeyProvider() string {
	return testPartitionKey
}

func createTestKinesisRecord(
	metrics int,
	data []byte,
) *kinesisRecord {

	partitionKey := testPartitionKey

	entry := &kinesis.PutRecordsRequestEntry{
		Data:            data,
		ExplicitHashKey: nil,
		PartitionKey:    &partitionKey,
	}

	return createKinesisRecord(entry, metrics)
}

func createTestKinesisRecords(
	data []byte,
	count int,
) []*kinesisRecord {

	partitionKey := testPartitionKey
	records := make([]*kinesisRecord, count, count)

	for i := 0; i < count; i++ {

		entry := &kinesis.PutRecordsRequestEntry{
			Data:            data,
			ExplicitHashKey: nil,
			PartitionKey:    &partitionKey,
		}

		records[i] = createKinesisRecord(entry, 1)
	}

	return records
}

func assertEndOfIterator(
	assert *assert.Assertions,
	iterator kinesisRecordIterator,
) {

	for i := 0; i < 2; i++ {

		record, err := iterator.Next()
		assert.NoError(err, "Next should not error")
		assert.Nil(record, "Next should not have returned another record")
	}
}
