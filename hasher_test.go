package bloomfilter

import (
	"crypto/md5"  // skipcq: GSC-G501
	"crypto/sha1" // skipcq: GSC-G505
	"fmt"
	"hash/crc64"
	"hash/fnv"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
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

// The package-level hashers are shared by every filter, and filters are read
// concurrently (rotate.Check only takes an RLock). A hash computed while
// another goroutine uses the same hasher must still match the sequential one,
// otherwise Check can miss an element that was added (a false negative) and
// Add can set the wrong bits, losing the element for good.
func TestHashersConcurrentUse(t *testing.T) {
	inputs := make([][]byte, 2000)
	for i := range inputs {
		inputs[i] = []byte(fmt.Sprintf("jti-0198f0c2-1a3b-7c4d-8e5f-%012d", i))
	}

	hashers := map[string]Hash{
		"md5":     MD5,
		"crc64":   CRC64,
		"sha1":    SHA1,
		"fnv64":   FNV64,
		"fnv128":  FNV128,
		"optimal": OptimalHashFactory(23)[0],
	}

	workers := runtime.GOMAXPROCS(0)
	if workers < 4 {
		workers = 4
	}

	for name, hash := range hashers {
		t.Run(name, func(t *testing.T) {
			want := make([][]uint, len(inputs))
			for i, in := range inputs {
				want[i] = hash(in)
			}

			var mismatches atomic.Int64
			var wg sync.WaitGroup
			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func(w int) {
					defer wg.Done()
					for round := 0; round < 20; round++ {
						for i := range inputs {
							j := (i + w*97) % len(inputs)
							if !reflect.DeepEqual(hash(inputs[j]), want[j]) {
								mismatches.Add(1)
							}
						}
					}
				}(w)
			}
			wg.Wait()

			if n := mismatches.Load(); n > 0 {
				t.Errorf("%d hashes differ from the sequential result under concurrent use", n)
			}
		})
	}
}

// Filters built before and after this change are merged and shipped over RPC,
// so the package hashers must keep producing exactly the same values.
func TestHashersMatchHashWrapper(t *testing.T) {
	wrapped := map[string][2]Hash{
		"md5":    {MD5, HashWrapper(md5.New())}, // skipcq: GO-S1023, GSC-G401
		"crc64":  {CRC64, HashWrapper(crc64.New(crc64.MakeTable(crc64.ECMA)))},
		"sha1":   {SHA1, HashWrapper(sha1.New())}, // skipcq: GO-S1025, GSC-G401
		"fnv64":  {FNV64, HashWrapper(fnv.New64())},
		"fnv128": {FNV128, HashWrapper(fnv.New128())},
	}
	for name, h := range wrapped {
		for _, in := range [][]byte{nil, {1, 2, 3}, []byte("jti-0198f0c2-1a3b-7c4d-8e5f-000000000001")} {
			if got, want := h[0](in), h[1](in); !reflect.DeepEqual(got, want) {
				t.Errorf("%s(%q) = %v, HashWrapper gives %v", name, in, got, want)
			}
		}
	}
}

// HashWrapper reuses the hash.Hash it is given, so it must serialize calls.
func TestHashWrapperConcurrentUse(t *testing.T) {
	hash := HashWrapper(fnv.New128())
	in := []byte("shared hash.Hash instance")
	want := hash(in)

	var mismatches atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20000; i++ {
				if !reflect.DeepEqual(hash(in), want) {
					mismatches.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	if n := mismatches.Load(); n > 0 {
		t.Errorf("%d hashes differ from the sequential result under concurrent use", n)
	}
}

// BenchmarkOptimalHashFactoryParallel measures the optimal hasher when many
// goroutines check a filter at once, as a gateway does.
func BenchmarkOptimalHashFactoryParallel(b *testing.B) {
	hash := OptimalHashFactory(23)[0]
	in := []byte("jti-0198f0c2-1a3b-7c4d-8e5f-000000000001")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			hash(in)
		}
	})
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

func TestDefaultHashFactory(t *testing.T) {
	for _, hash := range DefaultHashFactory(23) {
		array1 := []byte{1, 2, 3}
		if !reflect.DeepEqual(hash(array1), hash(array1)) {
			t.Error("undeterministic")
		}
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
