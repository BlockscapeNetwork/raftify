package raftify

import (
	"testing"
)

func TestHexToByte(t *testing.T) {
	if _, err := hexToByte(""); err != nil {
		t.Logf("Expected valid input, instead got: %v", err.Error())
		t.Fail()
	}
	if _, err := hexToByte("!§$%&/()="); err == nil {
		t.Logf("Expected error, instead got: %v", err.Error())
		t.Fail()
	}
	if _, err := hexToByte("123456789"); err == nil {
		t.Logf("Expected error, instead got %v", err.Error())
		t.Fail()
	}
	if _, err := hexToByte("1234567890"); err != nil {
		t.Logf("Expected valid input, instead got %v", err.Error())
		t.Fail()
	}
	if _, err := hexToByte("8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad"); err != nil {
		t.Logf("Expected valid input, instead got %v", err.Error())
		t.Fail()
	}
}
