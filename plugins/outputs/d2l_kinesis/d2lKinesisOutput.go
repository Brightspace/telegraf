package d2lkinesis

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const defaultMaxRecordRetries = 4

// Limits set by AWS (https://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecords.html)
const awsMaxRecordsPerRequest = 500
const awsMaxRecordSize = 1048576  // 1 MiB
const awsMaxRequestSize = 5242880 // 5 MiB

type (
	d2lKinesisOutput struct {

		// AWS Kinesis client configs
		Region      string `toml:"region"`
		AccessKey   string `toml:"access_key"`
		SecretKey   string `toml:"secret_key"`
		RoleARN     string `toml:"role_arn"`
		Profile     string `toml:"profile"`
		Filename    string `toml:"shared_credential_file"`
		Token       string `toml:"token"`
		EndpointURL string `toml:"endpoint_url"`

		// Stream configs
		MaxRecordRetries int    `toml:"max_record_retries"`
		MaxRecordSize    int    `toml:"max_record_size"`
		StreamName       string `toml:"stream_name"`

		// Internals
		Log             telegraf.Logger `toml:"-"`
		recordGenerator kinesisRecordGenerator
		serializer      serializers.Serializer
		svc             kinesisiface.KinesisAPI
	}
)

var sampleConfig = `
  ## Amazon REGION of kinesis endpoint.
  region = "ap-southeast-2"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## The maximum number of times to retry putting an individual Kinesis record
  # max_record_retries = 4

  ## The maximum Kinesis record size to put
  # max_record_size = 1048576

  ## Kinesis StreamName must exist prior to starting telegraf.
  stream_name = "StreamName"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

// SampleConfig returns the default configuration of the Processor
func (k *d2lKinesisOutput) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Processor
func (k *d2lKinesisOutput) Description() string {
	return "Configuration for the D2L AWS Kinesis output."
}

// Connect to the Output; connect is only called once when the plugin starts
func (k *d2lKinesisOutput) Connect() error {

	if k.StreamName == "" {
		return fmt.Errorf("stream_name is required")
	}

	if k.MaxRecordRetries < 0 {
		return fmt.Errorf("max_record_retries must be greater than or equal to 0")
	}

	if k.MaxRecordSize < 1000 {
		return fmt.Errorf("max_record_size must be greater than 1000 bytes")
	}

	if k.MaxRecordSize > awsMaxRecordSize {
		return fmt.Errorf("max_record_size must be less than or equal to the aws limit of %d bytes", awsMaxRecordSize)
	}

	generator, generatorErr := createGZipKinesisRecordGenerator(
		k.Log,
		k.MaxRecordSize,
		generateRandomPartitionKey,
		k.serializer,
	)
	if generatorErr != nil {
		return generatorErr
	}
	k.recordGenerator = generator

	credentialConfig := &internalaws.CredentialConfig{
		Region:      k.Region,
		AccessKey:   k.AccessKey,
		SecretKey:   k.SecretKey,
		RoleARN:     k.RoleARN,
		Profile:     k.Profile,
		Filename:    k.Filename,
		Token:       k.Token,
		EndpointURL: k.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()
	svc := kinesis.New(configProvider)

	_, err := svc.DescribeStreamSummary(&kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(k.StreamName),
	})
	k.svc = svc
	return err
}

// Close any connections to the Output. Close is called once when the output
// is shutting down. Close will not be called until all writes have finished,
// and Write() will not be called once Close() has been, so locking is not
// necessary.
func (k *d2lKinesisOutput) Close() error {
	return nil
}

// SetSerializer sets the serializer function for the interface.
func (k *d2lKinesisOutput) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

// Write takes in group of points to be written to the Output
func (k *d2lKinesisOutput) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}

	k.recordGenerator.Reset(metrics)

	return k.putRecordBatchesWithRetry(k.recordGenerator)
}

func (k *d2lKinesisOutput) putRecordBatchesWithRetry(
	recordIterator kinesisRecordIterator,
) error {

	attempt := 0
	for {

		failedRecords, err := k.putRecordBatches(recordIterator)
		if err != nil {
			return err
		}

		failedCount := len(failedRecords)
		if failedCount == 0 {
			return nil
		}

		attempt++
		if attempt > k.MaxRecordRetries {

			dropped := 0
			for _, record := range failedRecords {
				dropped += record.Metrics
			}

			k.Log.Errorf(
				"Unable to write %d record(s) to Kinesis after %d attempts; %d metrics dropped",
				failedCount,
				attempt,
				dropped,
			)

			return nil
		}

		k.Log.Debugf(
			"Retrying %d record(s)",
			failedCount,
		)
		recordIterator = createKinesisRecordSet(failedRecords...)
	}
}

func (k *d2lKinesisOutput) putRecordBatches(
	recordIterator kinesisRecordIterator,
) ([]*kinesisRecord, error) {

	batchRecordCount := 0
	batchRequestSize := 0
	var batch []*kinesisRecord

	var allFailedRecords []*kinesisRecord

	for {
		record, recordErr := recordIterator.Next()
		if recordErr != nil {
			return nil, recordErr
		}
		if record == nil {
			break
		}

		recordRequestSize := record.RequestSize
		if batchRequestSize+recordRequestSize > awsMaxRequestSize {

			failedRecords := k.putRecords(batch)
			allFailedRecords = append(allFailedRecords, failedRecords...)

			batchRecordCount = 0
			batchRequestSize = 0
			batch = nil
		}

		batchRecordCount++
		batchRequestSize += recordRequestSize
		batch = append(batch, record)

		if batchRecordCount >= awsMaxRecordsPerRequest {
			failedRecords := k.putRecords(batch)
			allFailedRecords = append(allFailedRecords, failedRecords...)

			batchRecordCount = 0
			batchRequestSize = 0
			batch = nil
		}
	}

	if batchRecordCount > 0 {
		failedRecords := k.putRecords(batch)
		allFailedRecords = append(allFailedRecords, failedRecords...)
	}

	return allFailedRecords, nil
}

// putRecords writes the set of records to Kinesis and returns the ones which failed
func (k *d2lKinesisOutput) putRecords(
	records []*kinesisRecord,
) []*kinesisRecord {

	totalRecordCount := len(records)

	entries := make([]*kinesis.PutRecordsRequestEntry, totalRecordCount)
	for i, record := range records {
		entries[i] = record.Entry
	}

	payload := kinesis.PutRecordsInput{
		Records:    entries,
		StreamName: aws.String(k.StreamName),
	}

	start := time.Now()
	resp, err := k.svc.PutRecords(&payload)
	duration := time.Since(start)

	if err != nil {

		k.Log.Warnf(
			"Unable to write %d records to Kinesis in %s: %s",
			totalRecordCount,
			duration.String(),
			err.Error(),
		)
		return records
	}

	successfulRecordCount := int64(totalRecordCount) - *resp.FailedRecordCount

	k.Log.Debugf(
		"Wrote %d of %d record(s) to Kinesis in %s",
		successfulRecordCount,
		totalRecordCount,
		duration.String(),
	)

	var failedRecords []*kinesisRecord

	failedRecordCount := *resp.FailedRecordCount
	if failedRecordCount > 0 {

		failedRecords = make([]*kinesisRecord, 0, failedRecordCount)

		for i, result := range resp.Records {
			if result.ErrorCode != nil {
				failedRecords = append(failedRecords, records[i])
			}
		}
	}

	return failedRecords
}

func init() {
	outputs.Add("d2l_kinesis", func() telegraf.Output {
		return &d2lKinesisOutput{

			MaxRecordRetries: defaultMaxRecordRetries,
			MaxRecordSize:    awsMaxRecordSize,
		}
	})
}
