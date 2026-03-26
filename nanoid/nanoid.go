package nanoid

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	DefaultAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	DefaultLength = 21
)

func init() {

}

// NewID generates a new ID using the default alphabet and length.
func NewID(len int) (string, error) {
	id, err := gonanoid.New(len)
	if err != nil {
		return "", err
	}

	return id, nil
}

// GenerateID generates a new ID using the default alphabet and length.
func GenerateID(len int) (string, error) {
	return Generate(DefaultAlphabet, len)
}

func Generate(alphabet string, len int) (string, error) {
	id, err := gonanoid.Generate(alphabet, len)
	if err != nil {
		return "", err
	}

	return id, nil
}
