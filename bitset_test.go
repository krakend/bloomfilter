package bloomfilter

import "testing"

func TestBitSet(t *testing.T) {
	callSet(t, NewBitSet(24))
}

func TestBitSet_Union(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := NewBitSet(24)

	callSet_Union(t, set1, set2)
}
