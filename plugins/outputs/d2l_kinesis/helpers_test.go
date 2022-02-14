package d2lkinesis

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
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

	entry := &types.PutRecordsRequestEntry{
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
		assert.Equal(io.EOF, err, "Next should return EOF error")
		assert.Nil(record, "Next should not have returned another record")
	}
}
