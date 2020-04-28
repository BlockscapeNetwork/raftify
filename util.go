package raftify

import (
	"encoding/hex"
)

// hexToByte decodes the string representation of the input into a byte slice.
func hexToByte(hexString string) ([]byte, error) {
	byteRep := make([]byte, hex.DecodedLen(len(hexString)))
	if _, err := hex.Decode(byteRep, []byte(hexString)); err != nil {
		return nil, err
	}
	return byteRep, nil
}
