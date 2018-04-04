package bloomfilter

import "math"

type Bloomfilter interface {
	Add([]byte)
	Check([]byte) bool
	Union(interface{}) (float64, error)
}

type Config struct {
	N        uint
	P        float64
	HashName string
}

var EmptyConfig = Config{
	N: 2,
	P: .5,
}

func M(n uint, p float64) uint {
	return uint(math.Ceil(-(float64(n) * math.Log(p)) / math.Log(math.Pow(2.0, math.Log(2.0)))))
}

func K(m, n uint) uint {
	return uint(math.Ceil(math.Log(2.0) * float64(m) / float64(n)))
}

type EmptySet int

func (e EmptySet) Check(_ []byte) bool                { return false }
func (e EmptySet) Add(_ []byte)                       {}
func (e EmptySet) Union(interface{}) (float64, error) { return -1, nil }
