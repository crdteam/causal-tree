package generate_uuid_elements

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// +-----------+
// | Utilities |
// +-----------+

// Provides a random MAC address.
func randomMAC() []byte {
	mac := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, mac); err != nil {
		panic(err.Error())
	}
	return mac
}

// Create UUIDv1, using local timestamp as lower bits and random MAC.
func RandomUUIDv1() uuid.UUID {
	uuid.SetNodeID(randomMAC())
	id, err := uuid.NewUUID()
	if err != nil {
		panic(fmt.Sprintf("creating UUIDv1: %v", err))
	}
	return id
}
