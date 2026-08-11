package bloom

import (
	"hash/maphash"
	"math"
)

var seed1 = maphash.MakeSeed()
var seed2 = maphash.MakeSeed()

type Filter struct {
	bits []uint64
	m    uint32
	k    uint32
}

func New(expectedN int, fpr float64) *Filter {
	m := uint32(math.Ceil(float64(expectedN) * math.Abs(math.Log(fpr)) / (math.Ln2 * math.Ln2)))
	k := uint32(math.Ceil((float64(m) / float64(expectedN)) * math.Ln2))
	if k < 1 {
		k = 1
	}
	words := (m + 63) / 64
	return &Filter{bits: make([]uint64, words), m: m, k: k}
}

func (f *Filter) Add(key []byte) {
	h1, h2 := f.hashes(key)
	for i := uint32(0); i < f.k; i++ {
		idx := (h1 + uint64(i)*h2 + uint64(i*i)) % uint64(f.m)
		f.bits[idx/64] |= 1 << (idx % 64)
	}
}

func (f *Filter) MayContain(key []byte) bool {
	h1, h2 := f.hashes(key)
	for i := uint32(0); i < f.k; i++ {
		idx := (h1 + uint64(i)*h2 + uint64(i*i)) % uint64(f.m)
		if f.bits[idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

func (f *Filter) hashes(key []byte) (uint64, uint64) {
	var h maphash.Hash
	h.SetSeed(seed1)
	h.Write(key)
	h1 := h.Sum64()
	h.SetSeed(seed2)
	h.Write(key)
	h2 := h.Sum64()
	if h2 == 0 {
		h2 = 1
	}
	return h1, h2
}

func (f *Filter) Cap() uint32  { return f.m }
func (f *Filter) Funcs() uint32 { return f.k }

func ExpectedFPR(m, k uint32, n int) float64 {
	km := float64(k) / float64(m)
	return math.Pow(1-math.Exp(-float64(n)*km), float64(k))
}
