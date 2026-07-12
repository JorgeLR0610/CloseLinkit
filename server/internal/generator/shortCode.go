package generator

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var ErrInvalidLength = errors.New("the provided length is invalid or too short")

const (
	base62alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	minLength      = 3
)

type ShortCodeGenerator struct {
	alphabet string
	length   int
}

func NewShortCodeGenerator(length int) (*ShortCodeGenerator, error) {
	if length < minLength {
		return nil, ErrInvalidLength
	}

	return &ShortCodeGenerator{
		alphabet: base62alphabet,
		length:   length,
	}, nil
}

func (sc *ShortCodeGenerator) GenerateShortCode() (string, error) {

	code := make([]byte, sc.length)

	for i := range sc.length {
		a, err := rand.Int(rand.Reader, big.NewInt(int64(len(sc.alphabet))))
		if err != nil {
			return "", err
		}

		code[i] = sc.alphabet[int(a.Int64())]
	}

	return string(code), nil
}
