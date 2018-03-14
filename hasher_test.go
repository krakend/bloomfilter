package bloomfilter

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHasher(t *testing.T) {
	for _, hash := range defaultHashers {
		array1 := []byte{1, 2, 3}
		if !reflect.DeepEqual(hash(array1), hash(array1)) {
			t.Error("undeterministic")
		}
	}
}

func BenchmarkHasher(b *testing.B) {
	for k, hash := range defaultHashers {
		b.Run(fmt.Sprintf("hasher %d", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				array1 := []byte{1, 2, 3}
				hash(array1)
			}
		})
	}
}

func TestOptimalHashFactory(t *testing.T) {
	for _, hash := range OptimalHashFactory(23) {
		array1 := []byte{1, 2, 3}
		if !reflect.DeepEqual(hash(array1), hash(array1)) {
			t.Error("undeterministic")
		}
	}
}

func BenchmarkOptimalHashFactory(b *testing.B) {
	for k, hash := range OptimalHashFactory(23) {
		b.Run(fmt.Sprintf("hasher %d", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				array1 := []byte{1, 2, 3}
				hash(array1)
			}
		})
	}
}

func CallSet(t *testing.T, set Set) {
	set.Add([]byte{1, 2, 3})
	if !set.Check([]byte{1, 2, 3}) {
		t.Error("failed check")
	}

	if set.Check([]byte{1, 2, 4}) {
		t.Error("unexpected check")
	}
}

func CallSetUnion(t *testing.T, set1, set2 Set) {
	elem := []byte{1, 2, 3}
	set1.Add(elem)
	if !set1.Check(elem) {
		t.Error("failed add set1 before union")
		return
	}

	if set2.Check(elem) {
		t.Error("unexpected check to union of set2")
		return
	}

	if _, err := set2.Union(set1); err != nil {
		t.Error("failed union set1 to set2", err.Error())
		return
	}

	if !set2.Check(elem) {
		t.Error("failed union check of set1 to set2")
		return
	}
}
