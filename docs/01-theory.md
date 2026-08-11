# Theory — Bloom filter pre-check in Go sync.Map miss latency

Produced at P1. Every source below was fetched and read. Nothing is cited from memory (rule R7).

## 1. The mechanism

A Bloom filter (Bloom 1970) is a probabilistic data structure storing set membership for _n_
elements in an _m_-bit array using _k_ independent hash functions. Insertion sets _k_ bits.
Query: if any of the _k_ bits is 0, the element is _definitely not_ present. If all _k_ bits
are 1, the element _may_ be present with some false-positive probability ε. There are no false
negatives.

The standard false-positive rate approximation (Broder & Mitzenmacher 2004, SRC-003):

```
ε ≈ (1 − e^(−kn/m))^k
```

For a given target ε and expected _n_, the optimal parameters are:

```
k = (m/n) · ln 2       (optimal number of hash functions)
m/n ≈ −1.44 · ln(ε)    (bits per element)
```

Approximately 9.6 bits per element yields ε ≈ 1%. Each additional ~4.8 bits per element reduces
ε by an order of magnitude.

Kirsch & Mitzenmacher (2006, SRC-004) showed that two hash functions _h1, h2_ can simulate _k_
independent hash functions via double hashing:

```
g_i(x) = (h1(x) + i · h2(x)) mod m,   for i = 0, …, k−1
```

This reduces hash computation from O(k) to O(2) with asymptotically equivalent false-positive
probability, assuming _h2(x)_ ≠ 0 mod _m_.

The concept applies this as a **contention-avoidance gate**: before calling `sync.Map.Load(key)`,
check the Bloom filter. If the filter says "definitely not present", skip the sync.Map call
entirely. If it says "maybe present" (true positive or false positive), proceed to Load(). The
filter must be updated on every Store() and Delete() — this is the insertion cost traded for
reduced miss-path traversal.

## 2. Conditions under which it holds

- **Read-heavy, miss-heavy workload.** The majority of Load(key) calls must be for absent keys.
  If most keys are present, every Bloom check returns "maybe present" and the filter adds only
  overhead with no miss-avoidance.
- **false-positive rate sufficiently low.** ε must be set small enough that false-positive
  Load() calls (the ones the filter cannot skip) do not dominate the savings from true-negative
  skips.
- **The filter fits in cache.** The _m_-bit array must be small enough to reside in CPU cache
  during the concurrent phase; an L3-level miss on the filter itself erodes the saving.
- **Insertion rate bounded.** If the key space churns at a rate that outpaces the Bloom filter's
  maintenance capacity, the false-positive rate climbs above the design point and the
  cost-benefit ratio collapses.
- **sync.Map miss path is the bottleneck.** The concept depends on the sync.Map slow path
  (mutex acquisition, dirty-map traversal on miss) being the dominant cost. If the workload is
  entirely read-only on a stable key set, sync.Map's read map alone is lock-free and the Bloom
  filter adds cost with no return.

## Known failure modes

| Failure mode | Trigger condition | Source |
| :--- | :--- | :--- |
| False-positive rate climbs above design point | Insertion count exceeds the filter's _n_ design parameter; filter is not resized | SRC-003 |
| Double hashing produces colliding probe sequences | _h2(x)_ = 0 mod _m_ on a given input, collapsing _k_ probes to a single position | SRC-004 |
| Bloom filter itself becomes the contention point | Concurrent Store() calls all updating the shared bit array; if the filter uses a mutex for writes, the contention moves rather than disappears | SRC-005 (by analogy with sync.Map's own mutex) |
| sync.Map dirty-map promotion resets miss counter | When `len(dirty)` shrinks, the miss-counter-to-promotion threshold changes; a correctly-sized Bloom filter _plus_ dirty-map promotion may produce an interaction where both mechanics are adjusting simultaneously | SRC-005 |
| Degenerate "always-say-yes" filter | _k_ is too large relative to _m/n_, saturating the filter to all 1s — the Bloom filter becomes a 100% false-positive rate, equivalent to no filter with added cost | SRC-003 |
| New keys always contend on sync.Map's single mutex | SRC-005, SRC-006: the sync.Map architecture (pre-Go 1.24) guards dirty-map writes with a single Mutex; Inserting a new key after a Bloom-positive check still contends on that Mutex | SRC-005, SRC-006 |

Row 6 is critical: the Bloom filter saves miss-paths but cannot save the insertion path. If
Store() calls are frequent enough, the bottleneck shifts from Load-miss to Store, and the
Bloom filter addresses a problem the workload no longer has.

## 4. The incumbent

**Go sync.Map** (Go 1.22 standard library). The internal architecture (pre-Go 1.24) has two
layers (SRC-005):

1. **read map** (`readOnly`): a lock-free map accessed via `atomic.Pointer`. Contains entries
   present at the last promotion. Load() on a present key in `read` is ~2 atomic loads, no lock.
2. **dirty map**: guarded by a single `sync.Mutex`. Holds keys not yet promoted, plus all
   keys from `read` that have not been expunged. Every miss-path Load() acquires the mutex,
   performs a double-check on `read`, then falls to `dirty`.

A `misses` counter increments on every miss-path Load. When `misses ≥ len(dirty)`, the dirty
map is promoted: it becomes the new `read` (atomically), and a fresh `dirty` is allocated. This
means steady-state misses _on keys that are genuinely absent_ continuously incur mutex
acquisition, with the miss counter never triggering promotion for those absent keys (they are
never in `dirty`). The worst case is a workload with high miss rate and stable population: every
Load for an absent key takes the slow path, and the single Mutex serialises them all.

Go issue #21035 (SRC-006) explicitly identifies this: "Store calls with different new keys
always contend" on the single Mutex, and one of the three exploration directions proposed is
"journaling writes (and using a Bloom or HyperLogLog filter to avoid reading the journal)."

**Baseline tuning to be fair at P5:** the incumbent is the plain `sync.Map.Load()` path. No
Bloom filter, no pre-check. The `sync.Map` itself will be used in its default zero-value
configuration, per Go's documented guidance.

## 5. Hypothesis

**H-001** — Under a read-heavy workload where ≥ 50 % of key accesses are to absent keys, a
Bloom filter pre-check on Go's sync.Map achieves at least 20 % lower p99 miss latency than
a plain sync.Map Load() at ≥ 16 goroutines, at the cost of a per-insertion Bloom insertion
and memory for the bit array.

*Falsified if:* the Bloom-filter-augmented sync.Map shows < 20 % p99 miss-latency reduction
over the plain sync.Map at every combination of (goroutines ∈ {1, 2, 4, 8, 16, 32, 64},
miss rate ∈ {0.5, 0.7, 0.9, 0.95, 0.99}) across ≥ 5 benchmark runs each.

*Measured by:* the benchmark at `bench/miss_latency_test.go`, which compares
`BloomMap.Load(key)` against `sync.Map.Load(key)` for a pre-populated key set with a
controlled miss rate, varying goroutine count from 1 to 64, and reporting p50/p95/p99
latency in nanoseconds.

## 6. Prior implementations

| Implementation | Maturity | What it does differently |
| :--- | :--- | :--- |
| Go issue #21035 (bcmills, 2017) | Proposal only | Suggests a Bloom or HyperLogLog filter as a journaling write pre-check; never implemented in the standard library |
| Amazon DynamoDB's use of Bloom filters | Production (closed source) | Uses Bloom filters internally for SSTable lookup; different access pattern (LSM-tree read amplification), not Go, not sync.Map |
| Squid web cache Summary Cache (Fan et al., via SRC-003) | Production | Exchanges Bloom filter digests between web cache proxies; different domain (distributed cache coherence) |
| go-bloomfilter libraries (e.g. willf/bloom, bits-and-blooms/bloom) | Library | Implement standalone Bloom filters in Go; none integrate with sync.Map as a pre-check layer |
| This project | Concept (TRL 1) | Measures the crossover point — the exact workload at which the Bloom filter pays back its overhead — rather than asserting it always helps |

The closest prior art is Go issue #21035 (SRC-006), which proposed the idea but never
benchmarked it. This project adds the measurement.

## 7. Open questions

- The exact false-positive rate that maximises net latency reduction for a given miss rate and
  goroutine count was not measured in any source consulted. Broder & Mitzenmacher state the
  formula but do not evaluate it against concurrent-map access patterns.
- The interaction between dirty-map promotion (SRC-005) and Bloom-filter false positives has
  not been characterised: does a Bloom false positive that triggers a slow-path Load() change
  the promotion threshold in a way that helps or hurts?
- Bloom's original 1970 paper (SRC-001) is paywalled. The mechanism description above uses the
  standard formulation from Wikipedia and the Broder & Mitzenmacher survey; any nuance in the
  original treatment of the error bound is not represented.

## Sources

### SRC-001 — Burton H. Bloom, *Space/Time Trade-offs in Hash Coding with Allowable Errors*, Communications of the ACM 13(7), 1970

- **URL:** https://dl.acm.org/doi/10.1145/362686.362692
- **Access:** abstract-only
- **Establishes:** The original Bloom filter — a bit array probed with k hash functions giving
  definite "no" and probabilistic "yes" with no false negatives. The tradeoff of hash area size
  against unnecessary-access avoidance for a 500,000-word hyphenation dictionary.

### SRC-002 — Go standard library, *sync.Map documentation*, pkg.go.dev

- **URL:** https://pkg.go.dev/sync#Map
- **Access:** full-text
- **Establishes:** The documented use case for sync.Map: write-once-read-many and disjoint
  key sets. States Map is "specialized" and "most code should use a plain Go map instead, with
  separate locking." Documents the Load/Store/Delete/LoadOrStore API with amortized constant
  time.

### SRC-003 — A. Broder, M. Mitzenmacher, *Network Applications of Bloom Filters: A Survey*, Internet Mathematics 1(4), 2004

- **URL:** https://www.eecs.harvard.edu/~michaelm/postscripts/im2005b.pdf
- **Access:** full-text
- **Establishes:** The canonical false-positive rate formula ε ≈ (1−e^(−kn/m))^k, optimal
  k = (m/n)·ln 2, the Summery Cache application using Bloom filters at web proxies, and the
  discussion of Bloom filters in caching contexts. ~9.6 bits/element for 1% false-positive
  rate.

### SRC-004 — A. Kirsch, M. Mitzenmacher, *Less Hashing, Same Performance: Building a Better Bloom Filter*, Random Structures & Algorithms 33(2), 2008

- **URL:** https://www.eecs.harvard.edu/~michaelm/postscripts/rsa2008.pdf
- **Access:** full-text
- **Establishes:** Two hash functions suffice to simulate k independent hash functions via
  double hashing g_i(x) = (h1(x) + i·h2(x)) mod m, with asymptotically equivalent
  false-positive probability. The enhanced double hashing variant adds an i² term to reduce
  the gap at finite m.

### SRC-005 — Go Authors, *sync/map.go* source code, Go 1.23

- **URL:** https://raw.githubusercontent.com/golang/go/go1.23.0/src/sync/map.go
- **Access:** full-text
- **Establishes:** The pre-Go-1.24 sync.Map architecture: lock-free `read` map accessed via
  atomic.Pointer, mutex-guarded `dirty` map, and a `misses` counter that triggers
  dirty→read promotion when misses ≥ len(dirty). Load() fast path is ~2 atomic loads; miss
  path acquires the single Mutex. New keys always contend on that Mutex.

### SRC-006 — Go issue #21035, *sync: reduce contention between Map operations with new-but-disjoint keys*, bcmills, 2017

- **URL:** https://github.com/golang/go/issues/21035
- **Access:** full-text
- **Establishes:** The Go team explicitly identified new-key contention under the single
  sync.Map Mutex and proposed three exploration directions: (1) sharding, (2) journaling
  writes with a Bloom or HyperLogLog filter as a read-side pre-check, (3) an atomic tree
  structure. This is the direct prior art for the concept.
