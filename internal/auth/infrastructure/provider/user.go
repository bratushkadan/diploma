package provider

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/bratushkadan/floral/pkg/resource"
)

const (
	UserIdByteLength = 8
	UserIdPrefix     = "i1"
)

func GenerateUserId() string {
	return resource.GenerateIdPrefix(UserIdByteLength, UserIdPrefix)
}

func ValidateUserId(id string) error {
	return resource.ValidateIdByteLenPrefix(id, UserIdByteLength, UserIdPrefix)
}

func Int64ToUserId(i int64) string {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, uint64(i))
	return resource.IdFromBytesPrefix(idBytes, UserIdPrefix)
}

func UserIdToInt64(id string) (int64, error) {
	bytes, err := resource.IdByteLenPrefixToBytes(id, UserIdByteLength, UserIdPrefix)
	if err != nil {
		return 0, fmt.Errorf("failed to convert user id to int64: %v", err)
	}

	if len(bytes) < 8 {
		return 0, errors.New("failed to convert user id to int64: user id too short")
	}

	return int64(binary.BigEndian.Uint64(bytes)), nil
}
