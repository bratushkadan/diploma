package resource

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

func IdFromBytesPrefix(idBytes []byte, prefix string) string {
	dst := make([]byte, base32.StdEncoding.EncodedLen(len(idBytes)))
	_, _ = rand.Read(dst)
	base32.StdEncoding.Encode(dst, idBytes)

	return prefix + string(bytes.TrimRight(bytes.ToLower(dst), "="))
}

func IdFromBytes(idBytes []byte) string {
	return IdFromBytesPrefix(idBytes, "")
}

func GenerateIdPrefix(byteLen int, prefix string) string {
	idBytes := make([]byte, byteLen)
	_, _ = rand.Read(idBytes)
	return IdFromBytesPrefix(idBytes, prefix)
}

func GenerateId(byteLen int) string {
	return GenerateIdPrefix(byteLen, "")
}

func IdByteLenPrefixToBytes(id string, byteLen int, prefix string) ([]byte, error) {
	if prefix != "" && !strings.HasPrefix(id, prefix) {
		return nil, errors.New("wrong id prefix")
	}

	id = strings.TrimPrefix(id, prefix)
	id = strings.ToUpper(id)
	idWithPadding := id + strings.Repeat("=", (8-len(id)%8)%8)

	decoded, err := base32.StdEncoding.DecodeString(idWithPadding)
	if err != nil {
		return nil, fmt.Errorf("failed to decode id: %w", err)
	}
	if byteLen != -1 && len(decoded) != byteLen {
		return nil, errors.New("wrong id length")
	}

	return decoded, nil
}
func ValidateIdByteLenPrefix(id string, byteLen int, prefix string) error {
	_, err := IdByteLenPrefixToBytes(id, byteLen, prefix)
	return err
}

func ValidateIdByteLen(id string, byteLen int) error {
	return ValidateIdByteLenPrefix(id, byteLen, "")
}

func ValidateIdPrefix(id string, prefix string) error {
	return ValidateIdByteLenPrefix(id, -1, prefix)
}

func ValidateId(id string) error {
	return ValidateIdByteLenPrefix(id, -1, "")
}
