package bloomfilter

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"testing"
	"time"
)

func TestRotate_keepRotating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dt := 5 * time.Millisecond

	rotate := &Rotate{
		Previous: NewBloomfilter(testCfg),
		Current:  NewBloomfilter(testCfg),
		Next:     NewBloomfilter(testCfg),
		Config:   testCfg,
		cancel:   cancel,
		mutex:    &sync.RWMutex{},
		TTL:      5,
		ctx:      ctx,
	}

	ch := make(chan time.Time)
	go rotate.keepRotating(ctx, ch)

	rotate.Add([]byte("test"))
	if !rotate.Check([]byte("test")) {
		t.Error("error: \"test\" not present")
	}

	serialized := new(bytes.Buffer)
	if err := gob.NewEncoder(serialized).Encode(rotate); err != nil {
		t.Errorf("error encoding Rotate, %s", err.Error())
	}

	ch <- time.Now()
	<-time.After(dt)
	if !rotate.Check([]byte("test")) {
		t.Error("error: \"test\" not present after 1 TTL")
	}

	ch <- time.Now()
	<-time.After(dt)
	if !rotate.Check([]byte("test")) {
		t.Error("error: \"test\" not present after 2 TTL")

	}
	ch <- time.Now()
	<-time.After(dt)
	if rotate.Check([]byte("test")) {
		t.Error("error: \"test\" present after 3 TTL")

	}

	rotate2 := new(Rotate)
	if err := gob.NewDecoder(serialized).Decode(rotate2); err != nil {
		t.Errorf("error encoding Rotate, %s", err.Error())
	}

	if !rotate2.Check([]byte("test")) {
		t.Error("error: \"test\" not present")
	}
}
