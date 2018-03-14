package rotate

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRotate_Union_ok(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	set2 := NewRotate(ctx, 5, testCfg)

	callSet_Union(t, set1, set2)
}

func TestRotate_Union_koIncorrectType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	set2 := 24

	if _, err := set1.Union(set2); err != ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleN(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	cfg := testCfg
	cfg.N = 1
	set2 := NewRotate(ctx, 5, cfg)
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: diferrent n values") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	cfg := testCfg
	cfg.P = 0.5
	set2 := NewRotate(ctx, 5, cfg)
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: diferrent p values") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleCurrentBFs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	set2 := NewRotate(ctx, 5, testCfg)
	set2.Current.k = 111
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "!= k2(111)") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatiblePreviousBFs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	set2 := NewRotate(ctx, 5, testCfg)
	set2.Previous.k = 111
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "!= k2(111)") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koDifferentHashFuncsBFs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	set2 := NewRotate(ctx, 5, testCfg)
	set2.Config.HashName = "optimal"
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: different hashers") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Unmarshal_okCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	elem := []byte("wwwww")
	set1.Add(elem)
	set2 := NewRotate(ctx, 5, testCfg)
	if set2.Check(elem) {
		t.Errorf("Unexpected elem %s in set2", elem)
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(set1); err != nil {
		t.Errorf("Unexpected error, %v", err)
	}
	if err := gob.NewDecoder(buf).Decode(set2); err != nil {
		t.Errorf("Unexpected error, %v", err)
	}
	if !set2.Check(elem) {
		t.Errorf("Expecting elem %s in set2", elem)
	}
}

func TestRotate_UnmarshalBinary_ko(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := NewRotate(ctx, 5, testCfg)
	if err := set1.UnmarshalBinary([]byte{}); err == nil {
		t.Error("should have given error")
	}
}

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
