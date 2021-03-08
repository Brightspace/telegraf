package d2lkinesis

import (
	"encoding/base64"

	"github.com/gofrs/uuid"
)

type partitionKeyGenerator func() string

func generateRandomPartitionKey() string {
	id, err := uuid.NewV4()
	if err != nil {
		return "default"
	}
	pk := base64.StdEncoding.EncodeToString(id.Bytes())
	return pk
}
