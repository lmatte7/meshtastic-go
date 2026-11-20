package gomesh

import (
	"crypto/rand"
)

// genPSK256 returns a random 256-bit pre-shared key for a channel.
func genPSK256() []byte {
	token := make([]byte, 32)
	// crypto/rand.Read never returns an error on supported platforms.
	rand.Read(token)
	return token
}
