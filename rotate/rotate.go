// Package rotate implemennts a sliding set of three bloomfilters: `previous`, `current` and `next` and the bloomfilter interface.
//
// When adding an element, it is stored in the `current` and `next` bloomfilters.
// When sliding (rotating), `current` passes to `previous` and `next` to `current`.
package rotate

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/devopsfaith/bloomfilter"
	"github.com/devopsfaith/bloomfilter/bloomfilter"
)

// New creates a new sliding set of 3 bloomfilters
// It uses a context and configuration
func New(ctx context.Context, cfg Config) *Bloomfilter {
	localCtx, cancel := context.WithCancel(ctx)
	prevCfg := bloomfilter.EmptyConfig
	prevCfg.HashName = cfg.HashName
	r := &Bloomfilter{
		Previous: baseBloomfilter.New(prevCfg),
		Current:  baseBloomfilter.New(cfg.Config),
		Next:     baseBloomfilter.New(cfg.Config),
		Config:   cfg,
		cancel:   cancel,
		mutex:    &sync.RWMutex{},
		ctx:      ctx,
	}

	go r.keepRotating(localCtx, time.NewTicker(time.Duration(cfg.TTL)*time.Second).C)
	return r
}

// Config contains a bloomfilter config and the rotation frequency TTL in sec
type Config struct {
	bloomfilter.Config
	TTL uint
}

// Bloomfilter type defines a sliding set of 3 bloomfilters
type Bloomfilter struct {
	Previous, Current, Next *baseBloomfilter.Bloomfilter
	Config                  Config
	mutex                   *sync.RWMutex
	ctx                     context.Context
	cancel                  context.CancelFunc
}

// Close sliding set of bloomfilters
func (bs *Bloomfilter) Close() {
	if bs != nil && bs.cancel != nil {
		bs.cancel()
	}
}

// Add element to sliding set of bloomfilters
func (bs *Bloomfilter) Add(elem []byte) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	bs.Next.Add(elem)
	bs.Current.Add(elem)
}

// Check if element in sliding set of bloomfilters
func (bs *Bloomfilter) Check(elem []byte) bool {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	return bs.Previous.Check(elem) || bs.Current.Check(elem)
}

// Union two sliding sets of bloomfilters
// Take care that false positive probability P,
// number of elements being filtered N and
// hashfunctions are the same
func (bs *Bloomfilter) Union(that interface{}) (float64, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	other, ok := that.(*Bloomfilter)
	if !ok {
		return bs.capacity(), bloomfilter.ErrImpossibleToTreat
	}
	if other.Config.N != bs.Config.N {
		return bs.capacity(), fmt.Errorf("error: diferrent n values %d vs. %d", other.Config.N, bs.Config.N)
	}

	if other.Config.P != bs.Config.P {
		return bs.capacity(), fmt.Errorf("error: diferrent p values %.2f vs. %.2f", other.Config.P, bs.Config.P)
	}

	if _, err := bs.Previous.Union(other.Previous); err != nil {
		return bs.capacity(), err
	}

	if _, err := bs.Current.Union(other.Current); err != nil {
		return bs.capacity(), err
	}

	if _, err := bs.Next.Union(other.Next); err != nil {
		return bs.capacity(), err
	}

	return bs.capacity(), nil
}

func (bs *Bloomfilter) keepRotating(ctx context.Context, c <-chan time.Time) {
	for {
		select {
		case <-c:
		case <-ctx.Done():
			return
		}

		bs.mutex.Lock()

		bs.Previous = bs.Current
		bs.Current = bs.Next
		bs.Next = baseBloomfilter.New(bloomfilter.Config{
			N:        bs.Config.N,
			P:        bs.Config.P,
			HashName: bs.Config.HashName,
		})

		bs.mutex.Unlock()
	}
}

// SerializibleBloomfilter used when (de)serializing a set of sliding bloomfilters
// It has exportable fields
type SerializibleBloomfilter struct {
	Previous, Current, Next *baseBloomfilter.Bloomfilter
	Config                  Config
}

// MarshalBinary serializes a set of sliding bloomfilters
func (bs *Bloomfilter) MarshalBinary() ([]byte, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	buf := new(bytes.Buffer)
	w := compressor.NewWriter(buf)
	err := gob.NewEncoder(w).Encode(SerializibleBloomfilter{
		Previous: bs.Previous,
		Next:     bs.Next,
		Current:  bs.Current,
		Config:   bs.Config,
	})

	w.Close()
	return buf.Bytes(), err
}

// MarshalBinary deserializes a set of sliding bloomfilters
func (bs *Bloomfilter) UnmarshalBinary(data []byte) error {
	if bs != nil && bs.cancel != nil {
		bs.cancel()

		bs.mutex.Lock()
		defer bs.mutex.Unlock()
	}

	buf := bytes.NewBuffer(data)
	r, err := compressor.NewReader(buf)
	if err != nil {
		return err
	}

	target := &SerializibleBloomfilter{}
	if err := gob.NewDecoder(r).Decode(target); err != nil && err != io.EOF {
		return err
	}

	ctx := context.Background()
	if bs != nil && bs.ctx != nil {
		ctx = bs.ctx
	}

	localCtx, cancel := context.WithCancel(ctx)

	*bs = Bloomfilter{
		Previous: target.Previous,
		Next:     target.Next,
		Current:  target.Current,
		Config:   target.Config,
		ctx:      ctx,
		cancel:   cancel,
		mutex:    new(sync.RWMutex),
	}

	go bs.keepRotating(localCtx, time.NewTicker(time.Duration(bs.Config.TTL)*time.Second).C)

	return nil
}

func (bs *Bloomfilter) capacity() float64 {
	return (bs.Previous.Capacity() + bs.Current.Capacity() + bs.Next.Capacity()) / 3.0
}
