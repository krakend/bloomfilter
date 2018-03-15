package rotate

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/letgoapp/go-bloomfilter"
	"github.com/letgoapp/go-bloomfilter/bfilter"
	"github.com/letgoapp/go-bloomfilter/testutils"
)

func TestRotate_Union_ok(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	set2 := New(ctx, 5, testutils.TestCfg)

	testutils.CallSetUnion(t, set1, set2)
}

func TestRotate_Union_koIncorrectType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	set2 := 24

	if _, err := set1.Union(set2); err != bloomfilter.ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleN(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	cfg := testutils.TestCfg
	cfg.N = 1
	set2 := New(ctx, 5, cfg)
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: diferrent n values") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	cfg := testutils.TestCfg
	cfg.P = 0.5
	set2 := New(ctx, 5, cfg)
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: diferrent p values") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koIncompatibleCurrentBFs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	set2 := New(ctx, 5, testutils.TestCfg2)
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: diferrent p values") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Union_koDifferentHashFuncsBFs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	set2 := New(ctx, 5, testutils.TestCfg)
	set2.Config.HashName = "optimal"
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "error: different hashers") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestRotate_Unmarshal_okCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	set1 := New(ctx, 5, testutils.TestCfg)
	elem := []byte("wwwww")
	set1.Add(elem)
	set2 := New(ctx, 5, testutils.TestCfg)
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

	set1 := New(ctx, 5, testutils.TestCfg)
	if err := set1.UnmarshalBinary([]byte{}); err == nil {
		t.Error("should have given error")
	}
}

func TestRotate_KeepRotating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dt := 5 * time.Millisecond

	rotate := &Bloomfilter{
		Previous: bfilter.New(testutils.TestCfg),
		Current:  bfilter.New(testutils.TestCfg),
		Next:     bfilter.New(testutils.TestCfg),
		Config:   testutils.TestCfg,
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

	rotate2 := new(Bloomfilter)
	if err := gob.NewDecoder(serialized).Decode(rotate2); err != nil {
		t.Errorf("error encoding Rotate, %s", err.Error())
	}

	if !rotate2.Check([]byte("test")) {
		t.Error("error: \"test\" not present")
	}
}
