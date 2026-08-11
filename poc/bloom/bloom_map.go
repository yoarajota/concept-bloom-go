package bloom

import "sync"

type BloomMap struct {
	m      sync.Map
	filter *Filter
}

func NewBloomMap(expectedN int, fpr float64) *BloomMap {
	return &BloomMap{filter: New(expectedN, fpr)}
}

func (bm *BloomMap) Load(key string) (any, bool) {
	if !bm.filter.MayContain([]byte(key)) {
		return nil, false
	}
	return bm.m.Load(key)
}

func (bm *BloomMap) Store(key string, val any) {
	bm.filter.Add([]byte(key))
	bm.m.Store(key, val)
}
