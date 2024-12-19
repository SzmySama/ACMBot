package helper

import "fmt"

// BitMap represents a set of bits.
type BitMap struct {
	bits []byte
	size int // Store the size of the BitMap
}

// NewBitMap creates a new BitMap with a specified size.
func NewBitMap(size int) *BitMap {
	return &BitMap{
		bits: make([]byte, (size+7)/8),
		size: size,
	}
}

// Set sets the bit at the given index to 1.
func (bm *BitMap) Set(index int) {
	if index < 0 || index >= bm.size {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, bm.size))
	}
	bm.bits[index/8] |= 1 << (uint(index) % 8)
}

// Clr clears the bit at the given index to 0.
func (bm *BitMap) Clr(index int) {
	if index < 0 || index >= bm.size {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, bm.size))
	}
	bm.bits[index/8] &^= 1 << (uint(index) % 8)
}

// Get returns whether the bit at the given index is set.
func (bm *BitMap) Get(index int) bool {
	if index < 0 || index >= bm.size {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, bm.size))
	}
	return bm.bits[index/8]&(1<<(uint(index)%8)) != 0
}
