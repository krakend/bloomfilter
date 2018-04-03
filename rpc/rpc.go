package rpc

import (
	"context"
	"fmt"

	"github.com/letgoapp/go-bloomfilter/rotate"
)

type Bloomfilter int

type AddInput struct {
	Elems [][]byte
}

type AddOutput struct {
	Count int
}

type CheckInput struct {
	Elems [][]byte
}

type CheckOutput struct {
	Checks []bool
}

type UnionInput struct {
	BF *rotate.Bloomfilter
}

type UnionOutput struct {
	Capacity float64
}

type Config struct {
	rotate.Config
	Port int
}

var (
	ErrNoBloomfilterInitialized = fmt.Errorf("Bloomfilter not initialized")
	bf                          *rotate.Bloomfilter
)

func (b *Bloomfilter) Add(in AddInput, out *AddOutput) error {
	if bf == nil {
		out.Count = 0
		return ErrNoBloomfilterInitialized
	}

	k := 0
	for _, elem := range in.Elems {
		bf.Add(elem)
		k++
	}
	out.Count = k

	return nil
}

func (b *Bloomfilter) Check(in CheckInput, out *CheckOutput) error {
	checkRes := make([]bool, len(in.Elems))

	if bf == nil {
		out.Checks = checkRes
		return ErrNoBloomfilterInitialized
	}

	for i, elem := range in.Elems {
		checkRes[i] = bf.Check(elem)
	}
	out.Checks = checkRes

	return nil
}

func (b *Bloomfilter) Union(in UnionInput, out *UnionOutput) error {
	if bf == nil {
		out.Capacity = 0
		return ErrNoBloomfilterInitialized
	}

	var err error
	out.Capacity, err = bf.Union(in.BF)

	return err
}

func New(ctx context.Context, cfg Config) *Bloomfilter {
	if bf != nil {
		bf.Close()
	}

	bf = rotate.New(ctx, cfg.Config)

	return new(Bloomfilter)
}

func (b *Bloomfilter) Close() {
	if bf != nil {
		bf.Close()
	}
}
