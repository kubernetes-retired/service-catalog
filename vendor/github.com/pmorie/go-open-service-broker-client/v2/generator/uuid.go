package generator

import (
	"fmt"
)

// IDFrom generates a UUID according to RFC 4122 based off a seed.
func IDFrom(seed string) string {
	uuid := make([]byte, 16)

	// Push the seed into the UUID.
	seedBytes := []byte(seed)
	for i := 0; i < 16; i++ {
		uuid[i] = seedBytes[i%len(seedBytes)]
	}

	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
