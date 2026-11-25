package bloomfilter

import (
	"crypto/md5"  // skipcq: GSC-G501
	"crypto/sha1" // skipcq: GSC-G505
	"encoding/binary"
	"fmt"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"sync"
)

type Hash func([]byte) []uint
type HashFactory func(uint) []Hash

const (
	HASHER_DEFAULT = "default"
	HASHER_OPTIMAL = "optimal"
)

var (
	defaultHashers = []Hash{
		MD5,
		CRC64,
		SHA1,
		FNV64,
		FNV128,
	}

	HashFactoryNames = map[string]HashFactory{
		HASHER_DEFAULT: DefaultHashFactory,
		HASHER_OPTIMAL: OptimalHashFactory,
	}

	ErrImpossibleToTreat = fmt.Errorf("unable to union")

	crc64Table = crc64.MakeTable(crc64.ECMA)

	// Each call gets its own hash.Hash, so these are safe for concurrent use:
	// filters share them and are read concurrently.
	MD5    = newHasher(md5.New)  // skipcq: GO-S1023, GSC-G401
	SHA1   = newHasher(sha1.New) // skipcq: GO-S1025, GSC-G401
	CRC64  = newHasher(func() hash.Hash { return crc64.New(crc64Table) })
	FNV64  = newHasher(fnv.New64)
	FNV128 = newHasher(fnv.New128)
)

func DefaultHashFactory(k uint) []Hash {
	if k > uint(len(defaultHashers)) {
		k = uint(len(defaultHashers))
	}
	return defaultHashers[:k]
}

func OptimalHashFactory(k uint) []Hash {
	return []Hash{
		func(b []byte) []uint {
			hs := FNV128(b)
			out := make([]uint, k)

			for i := range out {
				out[i] = hs[0] + uint(i)*hs[1]
			}
			return out
		},
	}
}

// HashWrapper turns h into a Hash. h is reused across calls, so calls are
// serialized to keep the returned Hash safe for concurrent use.
func HashWrapper(h hash.Hash) Hash {
	var mu sync.Mutex
	return func(elem []byte) []uint {
		mu.Lock()
		defer mu.Unlock()
		h.Reset()
		return sum(h, elem)
	}
}

// newHasher returns a Hash that builds a new hash.Hash with newHash on every
// call. Nothing is shared between calls, so it is safe for concurrent use.
func newHasher[H hash.Hash](newHash func() H) Hash {
	return func(elem []byte) []uint {
		return sum(newHash(), elem)
	}
}

// sum writes elem to h, which must be freshly created or reset, and returns
// the digest as little-endian 64-bit words.
func sum(h hash.Hash, elem []byte) []uint {
	h.Write(elem)
	result := h.Sum(nil)
	out := make([]uint, len(result)/8)
	for i := 0; i < len(result)/8; i++ {
		out[i] = uint(binary.LittleEndian.Uint64(result[i*8 : (i+1)*8]))
	}
	return out
}
