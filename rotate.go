package bloomfilter

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

func NewRotate(ctx context.Context, TTL uint, cfg Config) *Rotate {
	localCtx, cancel := context.WithCancel(ctx)
	r := &Rotate{
		Previous: NewBloomfilter(Config{
			N:        2,
			P:        .5,
			HashName: cfg.HashName}),
		Current: NewBloomfilter(cfg),
		Next:    NewBloomfilter(cfg),
		Config:  cfg,
		TTL:     TTL,

		cancel: cancel,
		mutex:  &sync.RWMutex{},
		ctx:    ctx,
	}

	go r.keepRotating(localCtx, time.NewTicker(time.Duration(TTL)*time.Second).C)
	return r
}

type Rotate struct {
	Previous, Current, Next *Bloomfilter
	TTL                     uint
	Config                  Config
	mutex                   *sync.RWMutex
	ctx                     context.Context
	cancel                  context.CancelFunc
}

func (bs *Rotate) Add(elem []byte) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	bs.Next.Add(elem)
	bs.Current.Add(elem)
}

func (bs *Rotate) Check(elem []byte) bool {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	return bs.Previous.Check(elem) || bs.Current.Check(elem)
}

func (bs *Rotate) Union(that interface{}) error {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	other, ok := that.(*Rotate)
	if !ok {
		return ErrImpossibleToTreat
	}
	if other.Config.N != bs.Config.N {
		return fmt.Errorf("error: diferrent n values %d vs. %d", other.Config.N, bs.Config.N)
	}

	if other.Config.P != bs.Config.P {
		return fmt.Errorf("error: diferrent p values %.2f vs. %.2f", other.Config.P, bs.Config.P)
	}

	hf0 := hashFactoryNames[bs.Config.HashName](bs.Next.k)
	hf1 := hashFactoryNames[other.Config.HashName](bs.Next.k)
	subject := make([]byte, 1000)
	rand.Read(subject)
	for i, f := range hf0 {
		if !reflect.DeepEqual(f(subject), hf1[i](subject)) {
			return errors.New("error: different hashers")
		}
	}

	if err := bs.Previous.Union(other.Previous); err != nil {
		return err
	}

	if err := bs.Current.Union(other.Current); err != nil {
		return err
	}

	return bs.Next.Union(other.Next)
}

func (bs *Rotate) keepRotating(ctx context.Context, c <-chan time.Time) {
	for {
		select {
		case <-c:
		case <-ctx.Done():
			return
		}

		bs.mutex.Lock()

		bs.Previous = bs.Current
		bs.Current = bs.Next
		bs.Next = NewBloomfilter(Config{
			N:        bs.Config.N,
			P:        bs.Config.P,
			HashName: bs.Config.HashName,
		})

		bs.mutex.Unlock()
	}
}

type SerializibleRotate struct {
	Previous, Current, Next *Bloomfilter
	Config                  Config
	TTL                     uint
}

func (bs *Rotate) MarshalBinary() ([]byte, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(SerializibleRotate{
		Previous: bs.Previous,
		Next:     bs.Next,
		Current:  bs.Current,
		Config:   bs.Config,
		TTL:      bs.TTL,
	})
	//zip buf.Bytes
	return buf.Bytes(), err
}

func (bs *Rotate) UnmarshalBinary(data []byte) error {
	if bs != nil && bs.cancel != nil {
		bs.cancel()

		bs.mutex.Lock()
		defer bs.mutex.Unlock()
	}
	//unzip data

	buf := bytes.NewBuffer(data)
	target := &SerializibleRotate{}

	if err := gob.NewDecoder(buf).Decode(target); err != nil {
		return err
	}

	ctx := context.Background()
	if bs != nil && bs.ctx != nil {
		ctx = bs.ctx
	}

	localCtx, cancel := context.WithCancel(ctx)

	*bs = Rotate{
		Previous: target.Previous,
		Next:     target.Next,
		Current:  target.Current,
		Config:   target.Config,
		TTL:      target.TTL,
		ctx:      ctx,
		cancel:   cancel,
		mutex:    new(sync.RWMutex),
	}

	go bs.keepRotating(localCtx, time.NewTicker(time.Duration(bs.TTL)*time.Second).C)

	return nil
}
