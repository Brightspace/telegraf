package d2lkinesis

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
)

func createKinesisRecord(
	entry *kinesis.PutRecordsRequestEntry,
	metrics int,
) *kinesisRecord {

	// Partition keys are included in the request size calculation.
	// This is assuming partition keys are ASCII.
	requestSize := len(entry.Data) + len(*entry.PartitionKey)

	return &kinesisRecord{
		Entry:       entry,
		Metrics:     metrics,
		RequestSize: requestSize,
	}
}

type kinesisRecord struct {

	// The AWS SDK PutRecords request entry
	Entry *kinesis.PutRecordsRequestEntry

	// The number of metrics serialized into the entry
	Metrics int

	// The PutRecords request size of the entry
	RequestSize int
}
