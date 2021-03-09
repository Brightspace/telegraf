package d2lkinesis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_kinesisRecordSet_Empty(t *testing.T) {
	assert := assert.New(t)

	var iterator kinesisRecordIterator
	iterator = createKinesisRecordSet([]*kinesisRecord{})

	assertEndOfIterator(assert, iterator)
}

func Test_kinesisRecordSet_SingleRecord(t *testing.T) {
	assert := assert.New(t)

	record1 := createTestKinesisRecord(1, []byte{0xa1})

	var iterator kinesisRecordIterator
	iterator = createKinesisRecordSet([]*kinesisRecord{
		record1,
	})

	record, err := iterator.Next()
	assert.NoError(err, "Next should not error")
	assert.Same(record1, record, "Should read record")

	assertEndOfIterator(assert, iterator)
}

func Test_kinesisRecordSet_MultipleRecords(t *testing.T) {
	assert := assert.New(t)

	record1 := createTestKinesisRecord(1, []byte{0xa1})
	record2 := createTestKinesisRecord(2, []byte{0xb2})
	record3 := createTestKinesisRecord(3, []byte{0xc3})

	var iterator kinesisRecordIterator
	iterator = createKinesisRecordSet([]*kinesisRecord{
		record1,
		record2,
		record3,
	})

	record, err := iterator.Next()
	assert.NoError(err, "Next should not error")
	assert.Same(record1, record, "Should read record 1")

	record, err = iterator.Next()
	assert.NoError(err, "Next should not error")
	assert.Same(record2, record, "Should read record 2")

	record, err = iterator.Next()
	assert.NoError(err, "Next should not error")
	assert.Same(record3, record, "Should read record 3")

	assertEndOfIterator(assert, iterator)
}
