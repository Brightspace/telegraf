package d2lkinesis

import (
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

func createKinesisRecord(
	entry *types.PutRecordsRequestEntry,
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
	Entry *types.PutRecordsRequestEntry

	// The number of metrics serialized into the entry
	Metrics int

	// The PutRecords request size of the entry
	RequestSize int
}
