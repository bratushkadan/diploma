package resource

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

func GenerateIdPrefix(byteLen int, prefix string) string {
	idBytes := make([]byte, byteLen)
	_, _ = rand.Read(idBytes)
	dst := make([]byte, base32.StdEncoding.EncodedLen(byteLen))
	_, _ = rand.Read(dst)
	base32.StdEncoding.Encode(dst, idBytes)

	return prefix + string(bytes.TrimRight(bytes.ToLower(dst), "="))
}

func GenerateId(byteLen int) string {
	return GenerateIdPrefix(byteLen, "")
}

func ValidateIdByteLenPrefix(id string, byteLen int, prefix string) error {
	if prefix != "" && !strings.HasPrefix(id, prefix) {
		return errors.New("wrong id prefix")
	}

	id = strings.TrimPrefix(id, prefix)
	id = strings.ToUpper(id)
	idWithPadding := id + strings.Repeat("=", (8-len(id)%8)%8)

	decoded, err := base32.StdEncoding.DecodeString(idWithPadding)
	if err != nil {
		return fmt.Errorf("failed to decode id: %w", err)
	}
	if byteLen != -1 && len(decoded) != byteLen {
		return errors.New("wrong id length")
	}
	return nil
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
