package kubeops

import (
	"math/rand"

	"github.com/google/uuid"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz123456789"
)

// NewUUIDString - generates a new UUID and converts it to string
func NewUUIDString() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

// RandString - generates random string with length
func RandString(length int) string {
	// One change made in Go 1.20 is that math/rand is now random by default.
	// Note rand.Seed(time.Now().UnixNano()) i
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
