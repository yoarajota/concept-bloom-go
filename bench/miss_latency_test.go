package bench

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/yoarajota/concept-bloom-go/src"
)

func missLatency(b *testing.B, nGoroutines, nKeys, nMisses int, useBloom bool) {
	keys := make([]string, nKeys)
	for i := 0; i < nKeys; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}
	missKeys := make([]string, nMisses)
	for i := 0; i < nMisses; i++ {
		missKeys[i] = fmt.Sprintf("miss-%d", i)
	}

	var bm *bloom.BloomMap
	var sm sync.Map
	if useBloom {
		bm = bloom.NewBloomMap(nKeys, 0.01)
		for i, k := range keys {
			bm.Store(k, i)
		}
	} else {
		for i, k := range keys {
			sm.Store(k, i)
		}
	}

	var writeCount atomic.Int64
	var done atomic.Bool
	go func() {
		for !done.Load() {
			k := fmt.Sprintf("churn-%d", writeCount.Add(1))
			if useBloom {
				bm.Store(k, 0)
				bm.Load(k)
			} else {
				sm.Store(k, 0)
				sm.Load(k)
			}
		}
	}()

	b.ResetTimer()
	b.SetParallelism(nGoroutines)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := missKeys[i%nMisses]
			i++
			if useBloom {
				bm.Load(k)
			} else {
				sm.Load(k)
			}
		}
	})
	done.Store(true)
}

func BenchmarkPlainMap_1g(b *testing.B)   { missLatency(b, 1, 100000, 100000, false) }
func BenchmarkBloomMap_1g(b *testing.B)   { missLatency(b, 1, 100000, 100000, true) }
func BenchmarkPlainMap_4g(b *testing.B)   { missLatency(b, 4, 100000, 100000, false) }
func BenchmarkBloomMap_4g(b *testing.B)   { missLatency(b, 4, 100000, 100000, true) }
func BenchmarkPlainMap_8g(b *testing.B)   { missLatency(b, 8, 100000, 100000, false) }
func BenchmarkBloomMap_8g(b *testing.B)   { missLatency(b, 8, 100000, 100000, true) }
func BenchmarkPlainMap_16g(b *testing.B)  { missLatency(b, 16, 100000, 100000, false) }
func BenchmarkBloomMap_16g(b *testing.B)  { missLatency(b, 16, 100000, 100000, true) }
func BenchmarkPlainMap_32g(b *testing.B)  { missLatency(b, 32, 100000, 100000, false) }
func BenchmarkBloomMap_32g(b *testing.B)  { missLatency(b, 32, 100000, 100000, true) }

func TestFPRConformance(t *testing.T) {
	n := 100000
	fpr := 0.01
	bf := bloom.New(n, fpr)
	for i := 0; i < n; i++ {
		bf.Add([]byte(fmt.Sprintf("key-%d", i)))
	}
	fp := 0
	probes := 100000
	for i := 0; i < probes; i++ {
		if bf.MayContain([]byte(fmt.Sprintf("miss-%d", i))) {
			fp++
		}
	}
	observed := float64(fp) / float64(probes)
	expected := bloom.ExpectedFPR(bf.Cap(), bf.Funcs(), n)
	t.Logf("expected FPR: %.4f, observed: %.4f (n=%d, m=%d, k=%d)",
		expected, observed, n, bf.Cap(), bf.Funcs())
	if observed > expected*2.0 {
		t.Errorf("observed FPR %.4f > 2x expected %.4f", observed, expected)
	}
}
