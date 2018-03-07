package bloomfilter

import (
	"fmt"

	"github.com/tmthrgd/go-bitset"
)

var ErrImpossibleToTreat = fmt.Errorf("unable to union")

type BitSet struct {
	bs     bitset.Bitset
	hasher Hash
}

func NewBitSet(m uint) *BitSet {
	return &BitSet{bitset.New(m), MD5}
}

func (bs *BitSet) Add(elem []byte) {
	bs.bs.Set(bs.hasher(elem)[0] % bs.bs.Len())
}

func (bs *BitSet) Check(elem []byte) bool {
	return bs.bs.IsSet(bs.hasher(elem)[0] % bs.bs.Len())
}

func (bs *BitSet) Union(that interface{}) error {
	other, ok := that.(*BitSet)
	if !ok {
		return ErrImpossibleToTreat
	}

	bs.bs.Union(bs.bs, other.bs)
	return nil
}
