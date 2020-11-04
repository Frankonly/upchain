package crypto

import "crypto/sha256"

func Hash(value []byte) []byte {
	hash := sha256.Sum256(value)
	return hash[:]
}

func HashNodes(left []byte, right []byte) []byte {
	return Hash(append(left, right...))
}
