package rotate

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/letgoapp/go-bloomfilter"
	bfilter "github.com/letgoapp/go-bloomfilter/bloomfilter"
)

func New(ctx context.Context, cfg Config) *Bloomfilter {
	localCtx, cancel := context.WithCancel(ctx)
	prevCfg := bloomfilter.EmptyConfig
	prevCfg.HashName = cfg.HashName
	r := &Bloomfilter{
		Previous: bfilter.New(prevCfg),
		Current:  bfilter.New(cfg.Config),
		Next:     bfilter.New(cfg.Config),
		Config:   cfg,
		cancel:   cancel,
		mutex:    &sync.RWMutex{},
		ctx:      ctx,
	}

	go r.keepRotating(localCtx, time.NewTicker(time.Duration(cfg.TTL)*time.Second).C)
	return r
}

type Config struct {
	bloomfilter.Config
	TTL uint
}

type Bloomfilter struct {
	Previous, Current, Next *bfilter.Bloomfilter
	Config                  Config
	mutex                   *sync.RWMutex
	ctx                     context.Context
	cancel                  context.CancelFunc
}

func (bs *Bloomfilter) Close() {
	if bs != nil && bs.cancel != nil {
		bs.cancel()
	}
}

func (bs *Bloomfilter) Add(elem []byte) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	bs.Next.Add(elem)
	bs.Current.Add(elem)
}

func (bs *Bloomfilter) Check(elem []byte) bool {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	return bs.Previous.Check(elem) || bs.Current.Check(elem)
}

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

	hf0 := bs.Next.HashFactoryNameK(bs.Config.HashName)
	hf1 := bs.Next.HashFactoryNameK(other.Config.HashName)
	subject := make([]byte, 1000)
	rand.Read(subject)
	for i, f := range hf0 {
		if !reflect.DeepEqual(f(subject), hf1[i](subject)) {
			return bs.capacity(), errors.New("error: different hashers")
		}
	}

	if _, err := bs.Previous.Union(other.Previous); err != nil {
		return bs.capacity(), err
	}

	if _, err := bs.Current.Union(other.Current); err != nil {
		return bs.capacity(), err
	}

	if _, err := bs.Current.Union(other.Next); err != nil {
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
		bs.Next = bfilter.New(bloomfilter.Config{
			N:        bs.Config.N,
			P:        bs.Config.P,
			HashName: bs.Config.HashName,
		})

		bs.mutex.Unlock()
	}
}

type SerializibleBloomfilter struct {
	Previous, Current, Next *bfilter.Bloomfilter
	Config                  Config
}

func (bs *Bloomfilter) MarshalBinary() ([]byte, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	err := gob.NewEncoder(w).Encode(SerializibleBloomfilter{
		Previous: bs.Previous,
		Next:     bs.Next,
		Current:  bs.Current,
		Config:   bs.Config,
	})

	w.Close()
	return buf.Bytes(), err
}

func (bs *Bloomfilter) UnmarshalBinary(data []byte) error {
	if bs != nil && bs.cancel != nil {
		bs.cancel()

		bs.mutex.Lock()
		defer bs.mutex.Unlock()
	}

	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
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
