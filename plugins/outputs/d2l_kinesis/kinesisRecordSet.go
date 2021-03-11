package d2lkinesis

import "io"

func createKinesisRecordSet(
	records ...*kinesisRecord,
) kinesisRecordIterator {

	return &kinesisRecordSet{
		count:   len(records),
		index:   0,
		records: records,
	}
}

type kinesisRecordSet struct {
	kinesisRecordIterator

	count   int
	index   int
	records []*kinesisRecord
}

func (s *kinesisRecordSet) Next() (*kinesisRecord, error) {

	index := s.index
	if index >= s.count {
		return nil, io.EOF
	}

	s.index++
	return s.records[index], nil
}
