package bfilter

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/letgoapp/go-bloomfilter"
	"github.com/tmthrgd/go-bitset"
)

type Bloomfilter struct {
	bs  bitset.Bitset
	m   uint
	k   uint
	h   []bloomfilter.Hash
	cfg bloomfilter.Config
}

func New(cfg bloomfilter.Config) *Bloomfilter {
	m := bloomfilter.M(cfg.N, cfg.P)
	k := bloomfilter.K(m, cfg.N)
	return &Bloomfilter{
		m:   m,
		k:   k,
		h:   bloomfilter.HashFactoryNames[cfg.HashName](k),
		bs:  bitset.New(m),
		cfg: cfg,
	}
}

func (b Bloomfilter) Add(elem []byte) {
	for _, h := range b.h {
		for _, x := range h(elem) {
			b.bs.Set(x % b.m)
		}
	}
}

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

func (b *Bloomfilter) Union(that interface{}) (float64, error) {
	other, ok := that.(*Bloomfilter)
	if !ok {
		return b.Capacity(), bloomfilter.ErrImpossibleToTreat
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

type SerializibleBloomfilter struct {
	BS       bitset.Bitset
	M        uint
	K        uint
	HashName string
	Cfg      bloomfilter.Config
}

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
		h:   bloomfilter.HashFactoryNames[target.HashName](target.K),
		cfg: target.Cfg,
	}

	return nil
}

func (bs *Bloomfilter) Capacity() float64 {
	return float64(bs.bs.Count()) / float64(bs.m)
}

func (bs *Bloomfilter) HashFactoryNameK(hashName string) []bloomfilter.Hash {
	return bloomfilter.HashFactoryNames[hashName](bs.k)
}
