package d2lkinesis

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateRandomPartitionKey(t *testing.T) {
	assert := assert.New(t)

	pk := generateRandomPartitionKey()
	assert.NotEmpty(pk, "Partition key should not be empty")

	pkBytes, decodeErr := base64.StdEncoding.DecodeString(pk)
	assert.NoError(decodeErr, "Partition key should be base64 string")
	assert.Len(pkBytes, 16, "Underlying partition key should be 16 bytes")
}
