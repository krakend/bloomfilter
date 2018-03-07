package bloomfilter

import "testing"

func TestBitSet(t *testing.T) {
	callSet(t, NewBitSet(24))
}

func TestBitSet_Union_ok(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := NewBitSet(24)

	callSet_Union(t, set1, set2)
}

func TestBitSet_Union_ko(t *testing.T) {
	set1 := NewBitSet(24)
	set2 := 24

	if err := set1.Union(set2); err != ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}
