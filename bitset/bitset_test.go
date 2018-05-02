package bitset

import (
	"testing"

	"github.com/devopsfaith/bloomfilter"
	"github.com/devopsfaith/bloomfilter/testutils"
)

func TestBitSet(t *testing.T) {
	testutils.CallSet(t, NewBitSet(24))
}

func TestBitSet_Union_ok(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := NewBitSet(24)

	testutils.CallSetUnion(t, set1, set2)
}

func TestBitSet_Union_ko(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := 24

	if _, err := set1.Union(set2); err != bloomfilter.ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}
