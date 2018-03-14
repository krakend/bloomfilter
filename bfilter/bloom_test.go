package bfilter

import (
	"bytes"
	"encoding/gob"
	"strings"
	"testing"
)

var testCfg = Config{
	N:        100,
	P:        0.001,
	HashName: "default",
}

func TestBloomfilter(t *testing.T) {
	bloomfilter.callSet(t, NewBloomfilter(testCfg))
}

func TestBloomfilter_Union_ok(t *testing.T) {
	set1 := NewBloomfilter(testCfg)
	set2 := NewBloomfilter(testCfg)

	bloomfilter.callSetUnion(t, set1, set2)
}

func TestBloomfilter_Union_koIncorrectType(t *testing.T) {
	set1 := NewBloomfilter(testCfg)
	set2 := 24

	if _, err := set1.Union(set2); err != ErrImpossibleToTreat {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestBloomfilter_Union_koDifferentM(t *testing.T) {
	set1 := NewBloomfilter(testCfg)
	set2 := NewBloomfilter(testCfg)
	set2.m = 111
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "!= m2(111)") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestBloomfilter_Union_koDifferentK(t *testing.T) {
	set1 := NewBloomfilter(testCfg)
	set2 := NewBloomfilter(testCfg)
	set2.k = 111
	if _, err := set1.Union(set2); err == nil || !strings.Contains(err.Error(), "!= k2(111)") {
		t.Errorf("Unexpected error, %v", err)
	}
}

func TestBloomfilter_gobEncoder(t *testing.T) {
	bf1 := NewBloomfilter(testCfg)
	bf1.Add([]byte("casa"))
	bf1.Add([]byte("grrrrr"))
	bf1.Add([]byte("something"))

	serialized := new(bytes.Buffer)
	if err := gob.NewEncoder(serialized).Encode(bf1); err != nil {
		t.Errorf("error encoding BF, %s", err.Error())
	}

	bf2 := new(Bloomfilter)
	if err := gob.NewDecoder(serialized).Decode(bf2); err != nil {
		t.Errorf("error encoding BF, %s", err.Error())
	}

	if !bf2.Check([]byte("casa")) {
		t.Error("error: \"casa\" not found")
	}
	if !bf2.Check([]byte("grrrrr")) {
		t.Error("error: \"grrrrr\" not found")
	}
	if !bf2.Check([]byte("something")) {
		t.Error("error: \"something\" not found")
	}
}

func TestUnmarshalBinary_ko(t *testing.T) {
	set1 := NewBloomfilter(testCfg)
	if err := set1.UnmarshalBinary([]byte{}); err == nil {
		t.Error("should have given error")
	}
}
