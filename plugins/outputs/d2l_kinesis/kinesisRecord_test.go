package d2lkinesis

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/assert"
)

func Test_createKinesisRecord(t *testing.T) {
	assert := assert.New(t)

	partitionKey := "test_partition_key"

	entry := kinesis.PutRecordsRequestEntry{
		Data:            []byte{0x01, 0x02, 0x03, 0x04},
		ExplicitHashKey: nil,
		PartitionKey:    &partitionKey,
	}

	record := createKinesisRecord(&entry, 8)
	assert.Equal(record.Entry, &entry)
	assert.Equal(record.Metrics, 8)
	assert.Equal(record.RequestSize, 22)
}
