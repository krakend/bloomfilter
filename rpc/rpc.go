package rpc

import "fmt"

type BFType int

type AddCheckArgs struct {
	elems [][]byte
}

type UnionArgs struct {
	bf *Bloomfilter
}

var (
	ErrNoBloomfilterInitialized = fmt.Errorf("Bloomfilter not initialized")
	bf                          *Bloomfilter
)

func (b BFType) Add(in AddCheckArgs, count *int) error {
	if bf == nil {
		*count = 0
		return ErrNoBloomfilterInitialized
	}

	k := 0
	for _, elem := range in.elems {
		in.bf.Add(elem)
		k++
	}
	*count = k

	return nil
}

func (b BFType) Check(in AddCheckArgs, checks *[]bool) error {
	checkRes := make([]bool, len(elems))

	if bf == nil {
		*checks = checkRes
		return ErrNoBloomfilterInitialized
	}

	for i, elem := range in.elems {
		checkRes[i] = in.bf.Check(elem)
	}
	*checks = checkRes

	return nil
}

func (b BFType) Union(in UnionArgs, count *float64) error {
	if bf == nil {
		*count = 0
		return ErrNoBloomfilterInitialized
	}

	var err error
	*count, err = in.bf1.Union(in.bf2)

	return err
}

func (b BFType) New(cfg Config) BFType {
	bf = NewBloomfilter(cfg)
}
