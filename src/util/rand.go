package util

import (
	"fmt"
	crand "crypto/rand"
)

func SecureRandom(bytes int) string {
	bs := make([]byte, bytes)
	if _, err := crand.Read(bs); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", bs)
}
