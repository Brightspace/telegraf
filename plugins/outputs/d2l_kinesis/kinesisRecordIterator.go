package d2lkinesis

type kinesisRecordIterator interface {
	Next() (*kinesisRecord, error)
}
