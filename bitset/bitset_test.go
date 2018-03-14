package bitset

import (
	"testing"

	bloomfilter "github.com/letgoapp/go-bloomfilter"
)

func TestBitSet(t *testing.T) {
	bloomfilter.CallSet(t, NewBitSet(24))
}

func TestBitSet_Union_ok(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := NewBitSet(24)

	bloomfilter.CallSetUnion(t, set1, set2)
}

func TestBitSet_Union_ko(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := 24

	if _, err := set1.Union(set2); err != ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}
