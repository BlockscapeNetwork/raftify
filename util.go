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

// List holding used ports in memory during package testing.
var portList []int

// reservePortskeeps a list for the port ranges used during package testing.
// It assignes a number of ports and returns the minimum and maximum values. Ports
// are reserved incrementally, starting from a base port. This way, the last element
// is always the highest port already in use.
// This should only be used for test functions in order to avoid port conflicts
// during package testing.
func reservePorts(number int) (ports []int) {
	nextPort := 3000
	if len(portList) != 0 {
		nextPort = portList[len(portList)-1] + 1
	}

	for i := 0; i < number; i++ {
		portList = append(portList, nextPort+i)
		ports = append(ports, nextPort+i)
	}
	return ports
}
