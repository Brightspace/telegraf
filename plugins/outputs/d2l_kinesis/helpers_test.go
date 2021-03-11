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
