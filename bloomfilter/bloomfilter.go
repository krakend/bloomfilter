package bloomfilter

import (
	"bytes"
	"encoding/gob"
	"fmt"

	bfilter "github.com/letgoapp/go-bloomfilter"
	"github.com/tmthrgd/go-bitset"
)

// Bloomfilter type
// It is defined by a bitset of m bits
// k hashfunctions and
// configuration
type Bloomfilter struct {
	bs  bitset.Bitset
	m   uint
	k   uint
	h   []bfilter.Hash
	cfg bfilter.Config
}

// New creates a new bloomfilter, from config
func New(cfg bfilter.Config) *Bloomfilter {
	m := bfilter.M(cfg.N, cfg.P)
	k := bfilter.K(m, cfg.N)
	return &Bloomfilter{
		m:   m,
		k:   k,
		h:   bfilter.HashFactoryNames[cfg.HashName](k),
		bs:  bitset.New(m),
		cfg: cfg,
	}
}

// Add element to bloomfilter
func (b Bloomfilter) Add(elem []byte) {
	for _, h := range b.h {
		for _, x := range h(elem) {
			b.bs.Set(x % b.m)
		}
	}
}

// Check element in bloomfilter
func (b Bloomfilter) Check(elem []byte) bool {
	for _, h := range b.h {
		for _, x := range h(elem) {
			if !b.bs.IsSet(x % b.m) {
				return false
			}
		}
	}
	return true
}

// Union of two bloomfilters
func (b *Bloomfilter) Union(that interface{}) (float64, error) {
	other, ok := that.(*Bloomfilter)
	if !ok {
		return b.Capacity(), bfilter.ErrImpossibleToTreat
	}

	if b.m != other.m {
		return b.Capacity(), fmt.Errorf("m1(%d) != m2(%d)", b.m, other.m)
	}

	if b.k != other.k {
		return b.Capacity(), fmt.Errorf("k1(%d) != k2(%d)", b.k, other.k)
	}

	b.bs.Union(b.bs, other.bs)

	return b.Capacity(), nil
}

// SerializibleBloomfilter used when (de)serializing a bloomfilter
// It has exportable fields
type SerializibleBloomfilter struct {
	BS       bitset.Bitset
	M        uint
	K        uint
	HashName string
	Cfg      bfilter.Config
}

// MarshalBinary serializes a bloomfilter
func (bs *Bloomfilter) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(&SerializibleBloomfilter{
		BS:       bs.bs,
		M:        bs.m,
		K:        bs.k,
		HashName: bs.cfg.HashName,
		Cfg:      bs.cfg,
	})
	//zip buf.Bytes

	return buf.Bytes(), err
}

// MarshalBinary deserializes a bloomfilter
func (bs *Bloomfilter) UnmarshalBinary(data []byte) error {
	//unzip data
	buf := bytes.NewBuffer(data)
	target := SerializibleBloomfilter{}

	if err := gob.NewDecoder(buf).Decode(&target); err != nil {
		return err
	}
	*bs = Bloomfilter{
		bs:  target.BS,
		m:   target.M,
		k:   target.K,
		h:   bfilter.HashFactoryNames[target.HashName](target.K),
		cfg: target.Cfg,
	}

	return nil
}

// Capacity returns the fill degree of the bloomfilter
func (bs *Bloomfilter) Capacity() float64 {
	return float64(bs.bs.Count()) / float64(bs.m)
}

// HashFactoryNameK returns the hashfunction given by name
func (bs *Bloomfilter) HashFactoryNameK(hashName string) []bfilter.Hash {
	return bfilter.HashFactoryNames[hashName](bs.k)
}
