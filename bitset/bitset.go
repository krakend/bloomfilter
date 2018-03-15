package bitset

import (
	"github.com/letgoapp/go-bloomfilter"
	"github.com/tmthrgd/go-bitset"
)

type BitSet struct {
	bs     bitset.Bitset
	hasher bloomfilter.Hash
}

func NewBitSet(m uint) *BitSet {
	return &BitSet{bitset.New(m), bloomfilter.MD5}
}

func (bs *BitSet) Add(elem []byte) {
	bs.bs.Set(bs.hasher(elem)[0] % bs.bs.Len())
}

func (bs *BitSet) Check(elem []byte) bool {
	return bs.bs.IsSet(bs.hasher(elem)[0] % bs.bs.Len())
}

func (bs *BitSet) Union(that interface{}) (float64, error) {
	other, ok := that.(*BitSet)
	if !ok {
		return bs.getCount(), bloomfilter.ErrImpossibleToTreat
	}

	bs.bs.Union(bs.bs, other.bs)
	return bs.getCount(), nil
}

func (bs *BitSet) getCount() float64 {
	return float64(bs.bs.Count()) / float64(bs.bs.Len())
}
