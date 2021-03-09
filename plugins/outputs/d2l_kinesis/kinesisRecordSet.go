package d2lkinesis

func createKinesisRecordSet(
	records []*kinesisRecord,
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
		return nil, nil
	}

	s.index++
	return s.records[index], nil
}
