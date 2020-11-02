package upchain

import (
	"fmt"
	"math/bits"
)

// ref Libra Position module

// maxLevel for index in uint64
const maxLevel = 63

// InorderIndex represents the inorder traversal index of a binary tree with limited level
type InorderIndex uint64

// FromLeavesOnLevel calculates inorder index from the number of leaves upper certain level
func FromLeavesOnLevel(leavesOnLevel uint64, level int) InorderIndex {
	return InorderIndex(leavesOnLevel<<(level+1) | (1<<level - 1))
}

// FromLeaves calculates inorder index from the number of leaves
func FromLeaves(leaves uint64) InorderIndex {
	return FromLeavesOnLevel(leaves, 0)
}

// FromPostorder calculates inorder index from postorder index
func FromPostorder(postorder uint64) InorderIndex {
	bitmap := uint64(0)
	fullBinarySize := ^uint64(0)

	for i := maxLevel; i >= 0; i-- {
		if postorder >= fullBinarySize {
			postorder -= fullBinarySize
			bitmap |= (1 << i)
		}
		fullBinarySize >>= 1
	}

	return FromLeavesOnLevel(bitmap>>postorder, int(postorder))
}

func (i InorderIndex) children() uint64 {
	return uint64(isolateRightMostZeroBit(i))<<1 - 2
}

// Postorder returns the postorder index converted from inorder index
func (i InorderIndex) Postorder() uint64 {
	onesUpToLevel := uint64(isolateRightMostZeroBit(i)) - 1
	unsetLevelZeros := uint64(i) ^ onesUpToLevel
	return i.children() + unsetLevelZeros - uint64(bits.OnesCount64(unsetLevelZeros))
}

// Parent returns the parent
func (i InorderIndex) Parent() InorderIndex {
	return (i | isolateRightMostZeroBit(i)) & ^(isolateRightMostZeroBit(i) << 1)
}

// Sibling returns the sibling
func (i InorderIndex) Sibling() InorderIndex {
	return i ^ (isolateRightMostZeroBit(i) << 1)
}

//Level calculates the level of inorder index
func (i InorderIndex) Level() int {
	return bits.TrailingZeros64(^uint64(i))
}

// LeavesOnLevel returns n that i is the n-th leaf on this level
func (i InorderIndex) LeavesOnLevel() uint64 {
	return uint64(i) >> (1 + i.Level())
}

// IsLeaf judges whether the inorder index is a leaf
func (i InorderIndex) IsLeaf() bool {
	return i&1 == 0
}

// IsLeftChild judges whether the inorder index is or can be a left child
func (i InorderIndex) IsLeftChild() bool {
	return i&isolateRightMostZeroBit(i)<<1 == 0
}

// IsRightChild judges whether the inorder index is or can be a right child
func (i InorderIndex) IsRightChild() bool {
	return !i.IsLeftChild()
}

// LeftChild returns the left child
func (i InorderIndex) LeftChild() (InorderIndex, error) {
	if !i.IsLeaf() {
		return 0, fmt.Errorf("Not leaf")
	}
	return i & ^(isolateRightMostZeroBit(i) >> 1), nil
}

// RightChild returns the right child
func (i InorderIndex) RightChild() (InorderIndex, error) {
	if !i.IsLeaf() {
		return 0, fmt.Errorf("Not leaf")
	}
	return (i | isolateRightMostZeroBit(i)) & ^(isolateRightMostZeroBit(i) >> 1), nil
}

// LeftMostChild returns the left-most child
func (i InorderIndex) LeftMostChild() InorderIndex {
	level := i.Level()
	return (i >> level) << level
}

// RightMostChild returns the right-most child
func (i InorderIndex) RightMostChild() InorderIndex {
	return i + (InorderIndex(i.children()) >> 1)
}

// RootLevelFromLeaves calculates the root level of a binary tree containing certain number of leaves
func RootLevelFromLeaves(leaves uint64) int {
	return maxLevel + 1 - bits.LeadingZeros64(leaves-1)
}

func isolateRightMostZeroBit(x InorderIndex) InorderIndex {
	return (^x) & (x + 1)
}
