package gomesh

import (
	"math/rand"
	"strconv"
	"time"
)

// convPSK converts user input into a preset value for the PSK
func convPSK(param string) (psk []byte, err error) {
	psk = []byte{0x01}
	if param == "random" {
		psk = genPSK256()
	} else if param == "none" {
		psk = []byte{0x00}
	} else if param == "default" {
		psk = []byte{0x01}
	} else if param[:6] == "simple" {
		// Use one of the single byte encodings
		simVal, err := strconv.Atoi(param[6:])
		if err != nil {
			return nil, err
		}
		bVal := byte(simVal + 1)
		psk = []byte{bVal}
	}

	return
}

func genPSK256() []byte {

	token := make([]byte, 32)
	rand.Seed(time.Now().UnixNano())
	rand.Read(token)

	return token
}
