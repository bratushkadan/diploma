package idhash

import (
	"errors"
	"fmt"
	"strings"

	"github.com/speps/go-hashids/v2"
)

const (
	AlphabetAlphanumericLowercase = "abcdefghijklmnopqrstuvwxyz0123456789"
	AlphabetAlphanumericUppercase = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	ErrInvalidIdPrefix = errors.New("invalid id prefix")
)

type IdHasher struct {
	h      *hashids.HashID
	prefix string
}
type idHasherConf struct {
	hd     *hashids.HashIDData
	prefix string
}

type Option = func(*idHasherConf)

func WithMinLen(minLen int) Option {
	return func(conf *idHasherConf) {
		conf.hd.MinLength = minLen
	}
}
func WithAlphabet(alphabet string) Option {
	return func(conf *idHasherConf) {
		conf.hd.Alphabet = alphabet
	}
}

// Encode ids with the provided prefix.
// Validate ids for starting with the provided prefix before decoding.
func WithPrefix(prefix string) Option {
	return func(conf *idHasherConf) {
		conf.prefix = prefix
	}
}

var defaultOpts = []Option{WithMinLen(12), WithAlphabet(AlphabetAlphanumericLowercase)}

func New(salt string, opts ...Option) (IdHasher, error) {
	hd := hashids.NewData()
	hd.Salt = salt

	conf := &idHasherConf{
		hd: hd,
	}

	for _, opt := range defaultOpts {
		opt(conf)
	}
	for _, opt := range opts {
		opt(conf)
	}

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return IdHasher{}, fmt.Errorf("failed to setup hashids hasher: %v", err)
	}

	return IdHasher{h: h}, nil
}

func (h *IdHasher) Encode(data []int) (string, error) {
	encoded, err := h.h.Encode(data)
	if err != nil {
		return "", err
	}
	return h.prefix + encoded, nil
}
func (h *IdHasher) EncodeInt(data int) (string, error) {
	return h.Encode([]int{data})

}
func (h *IdHasher) EncodeInt64(data int64) (string, error) {
	return h.EncodeInt(int(data))

}
func (h *IdHasher) EncodeUint64(data uint64) (string, error) {
	return h.EncodeInt(int(data))
}

func (h *IdHasher) Decode(s string) ([]int, error) {
	var ok bool
	s, ok = strings.CutPrefix(s, h.prefix)
	if !ok {
		return nil, ErrInvalidIdPrefix
	}

	decoded, err := h.h.DecodeWithError(s)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
func (h *IdHasher) DecodeInt(s string) (int, error) {
	decoded, err := h.Decode(s)
	if err != nil || len(decoded) == 0 {
		return 0, err
	}
	return decoded[0], nil
}
func (h *IdHasher) DecodeInt64(s string) (int64, error) {
	decoded, err := h.Decode(s)
	if err != nil || len(decoded) == 0 {
		return 0, err
	}
	return int64(decoded[0]), nil
}
func (h *IdHasher) DecodeUint64(s string) (uint64, error) {
	decoded, err := h.Decode(s)
	if err != nil || len(decoded) == 0 {
		return 0, err
	}
	return uint64(decoded[0]), nil
}
