package upchain

// Digester can digest a merkle tree
type Digester interface {
}

// MerkleTree represents the basic merkle tree functions
type MerkleTree interface {
	CalculateHash() []byte
	Append([]byte)
	//Equals(MerkleTree) bool
}

// SimpleMerkleTree represents a appendable merkle binary tree
type SimpleMerkleTree struct {
}

// Node represents basic structure of node of merkle tree, who implements the interface MerkleTree
type Node struct {
	Leaves     []*Node
	Appendable bool
}
