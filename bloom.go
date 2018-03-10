package bloomfilter

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/tmthrgd/go-bitset"
)

type Set interface {
	Add([]byte)
	Check([]byte) bool
	Union(interface{}) error
}

type Config struct {
	N        uint
	P        float64
	HashName string
}

type Bloomfilter struct {
	bs  bitset.Bitset
	m   uint
	k   uint
	h   []Hash
	cfg Config
}

func M(n uint, p float64) uint {
	return uint(math.Ceil(-(float64(n) * math.Log(p)) / math.Log(math.Pow(2.0, math.Log(2.0)))))
}

func K(m, n uint) uint {
	return uint(math.Ceil(math.Log(2.0) * float64(m) / float64(n)))
}

func NewBloomfilter(cfg Config) *Bloomfilter {
	m := M(cfg.N, cfg.P)
	k := K(m, cfg.N)
	return &Bloomfilter{
		m:   m,
		k:   k,
		h:   hashFactoryNames[cfg.HashName](k),
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

func (b *Bloomfilter) Union(that interface{}) error {
	other, ok := that.(*Bloomfilter)
	if !ok {
		return ErrImpossibleToTreat
	}

	if b.m != other.m {
		return fmt.Errorf("m1(%d) != m2(%d)", b.m, other.m)
	}

	if b.k != other.k {
		return fmt.Errorf("k1(%d) != k2(%d)", b.k, other.k)
	}

	b.bs.Union(b.bs, other.bs)
	return nil
}

type SerializibleBloomfilter struct {
	BS       bitset.Bitset
	M        uint
	K        uint
	HashName string
	Cfg      Config
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
		h:   hashFactoryNames[target.HashName](target.K),
		cfg: target.Cfg,
	}

	return nil
}
