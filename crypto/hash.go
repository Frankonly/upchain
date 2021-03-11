package crypto

import "crypto/sha256"

// Hash hashes bytes by SHA256
func Hash(value []byte) []byte {
	hash := sha256.Sum256(value)
	return hash[:]
}

// HashNodes hashes two nodes into one
func HashNodes(left []byte, right []byte) []byte {
	return Hash(append(left, right...))
}
