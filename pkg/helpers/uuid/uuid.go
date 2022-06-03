package uuid

import (
	"fmt"
	"github.com/google/uuid"
)

// Generator represents a UUID generator.
type Generator interface {
	Make() (string, error)
}

// UUIDGen is a concrete implementation of Generator interface.
type UUIDGen struct{}

func New() Generator {
	return &UUIDGen{}
}

// Make generates a new random UUID string .
func (u *UUIDGen) Make() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("id_%X%X", uid[0:4], uid[4:6]), nil
}
