// Package bloomfilter contains common data and interfaces needed to implement bloomfilters.
//
// It is based on the theory explained in: http://llimllib.github.io/bloomfilter-tutorial/
// In the repo, there are created the following types of bloomfilter: derived from bitset, sliding bloomfilters
// and rpc bloomfilter implementation.
package bloomfilter

import "math"

// Bloomfilter interface implemented in the different packages
type Bloomfilter interface {
	Add([]byte)
	Check([]byte) bool
	Union(interface{}) (float64, error)
}

// Config for bloomfilter defining the parameters:
// P - desired false positive probability, N - number of elements to be stored in the filter and
// HashName - the name of the particular hashfunction
type Config struct {
	N        uint
	P        float64
	HashName string
}

// EmptyConfig configuration used for first empty `previous` bloomfilter in the sliding three bloomfilters
var EmptyConfig = Config{
	N: 2,
	P: .5,
}

// M function computes the length of the bit array of the bloomfilter as function of n and p
func M(n uint, p float64) uint {
	return uint(math.Ceil(-(float64(n) * math.Log(p)) / math.Log(math.Pow(2.0, math.Log(2.0)))))
}

// K function computes the number of hashfunctions of the bloomfilter as function of n and p
func K(m, n uint) uint {
	return uint(math.Ceil(math.Log(2.0) * float64(m) / float64(n)))
}

// EmptySet type is a synonym of int
type EmptySet int

// Check implementation for EmptySet
func (e EmptySet) Check(_ []byte) bool { return false }

// Add implementation for EmptySet
func (e EmptySet) Add(_ []byte) {}

// Union implementation for EmptySet
func (e EmptySet) Union(interface{}) (float64, error) { return -1, nil }
