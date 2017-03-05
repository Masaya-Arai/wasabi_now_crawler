package util

import "testing"

func TestSecureRandom(t *testing.T) {
	result := SecureRandom(16)
	if len(result) != 32 {
		t.Fatal("util.SecureRandom returns value of invalid length.")
	}
}

func TestSecureRandom2(t *testing.T) {
	result := SecureRandom(16)
	for i := 0; i < 10000; i++ {
		v := SecureRandom(16)
		if v == result {
			t.Fatal("util.SecureRandom returns same values.")
		}
	}
}