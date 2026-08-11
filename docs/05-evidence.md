# Evidence ledger — Bloom filter pre-check in Go sync.Map: crossover point for miss latency under read-heavy load

Every claim in this repository resolves to an entry here (rule R3). An entry without a
reproduction command is a note, not evidence, and fails validation.

**Required in every entry:** a fenced command block, an `Environment:` line, and a `Result:`
line. Headings must be exactly `### E-###  —  <title>` so the tooling can parse them.

IDs are never reused. Evidence that stops reproducing is marked `Status: broken` and the
readiness level that depended on it comes down (rule R1).

---

### E-001 — Literature survey

**Claim.** The basic principles of Bloom filters and Go sync.Map are understood and documented
from primary sources (docs/01-theory.md). TRL 1.

**Environment:** Any machine with a web browser. No runtime.

```bash
# The search that produced the source list below:
# 1. Searched: "bloom filter optimal k formula" → Broder & Mitzenmacher (SRC-003)
# 2. Searched: "golang sync.map internal architecture" → Go source map.go (SRC-005)
# 3. Searched: "golang sync.map miss latency issue 21035" → Go issue #21035 (SRC-006)
# 4. Searched: "less hashing same performance" → Kirsch & Mitzenmacher (SRC-004)
# 5. Searched: Go sync.Map docs at pkg.go.dev → sync.Map public API (SRC-002)
# 6. Searched: "Bloom 1970 space time tradeoffs hash coding" → ACM page (SRC-001)
grep -c "SRC-00" docs/01-theory.md
```

**Result:** Six sources documented: 1 abstract-only, 5 full-text. The grep returns ≥ 6
(matches SRC-001 through SRC-006 source headings). Run date: 2026-08-10.

**Status:** reproducing
**Supports:** TRL 1 for `core`
**Recorded:** 2026-08-10

---

### E-002 — PoC: Bloom filter false-positive rate conformance

**Claim.** The Bloom filter implementation conforms to the analytical false-positive rate
formula ε ≈ (1−e^(−kn/m))^k (SRC-003), and the BloomMap wraps sync.Map with a pre-check
that avoids Load() on absent keys (docs/01-theory.md § 1). TRL 2–3.

**Environment:** Go 1.22.2, Linux amd64, 13th Gen Intel i7-13650HX.

```bash
go test ./bench/... -run TestFPR -v
```

**Result:** FPR observed 0.0101 vs expected 0.0100 for n=100,000, m=958,506, k=7
(within 1% of the predicted value). The conformance test passes — the implementation's
FPR stays within 2x of the analytical bound. ISO 5055 quality gates exit 0 (make quality:
go vet, staticcheck, gitleaks, govulncheck). Architecture decisions documented in D-001.

**Status:** reproducing
**Supports:** H-001, S-001, TRL 4 for `core`
**Recorded:** 2026-08-10

---

### E-003 — Benchmark: BloomMap vs plain sync.Map miss latency

**Claim.** The Bloom filter pre-check does not reduce miss latency compared to plain sync.Map
under read-heavy concurrent access across 1–32 goroutines with a background writer maintaining
the dirty map. Hypothesis H-001 is falsified.

**Environment:** Go 1.22.2, Linux amd64, 13th Gen Intel i7-13650HX, 100k pre-populated keys,
1% Bloom filter false-positive rate (m=958,506, k=7), background writer inserting keys
continuously.

```bash
go test ./bench/... -bench=. -benchmem -count=5 -benchtime=1s
```

**Result:** Mean ns/op across 5 runs:

| Goroutines | Plain sync.Map | BloomMap | Bloom / Plain |
| :--- | :-: | :-: | :-: |
| 1 | 61.74 ns | 62.28 ns | +0.9% (slower) |
| 4 | 23.39 ns | 27.59 ns | +18.0% (slower) |
| 8 | 18.07 ns | 22.52 ns | +24.6% (slower) |
| 16 | 17.01 ns | 18.11 ns | +6.5% (slower) |
| 32 | 15.89 ns | 18.68 ns | +17.6% (slower) |

The Bloom filter pre-check is consistently slower than plain sync.Map at every goroutine
count tested. The hypothesis predicted ≥20% improvement at ≥16 goroutines; the observed
result is a 6–25% penalty.

**Analysis:** sync.Map's internal miss path (Go 1.22) is already highly optimised: atomic
reads on the `read` map, and the mutex is only acquired when a dirty map exists. Even with a
background writer creating a dirty map, the per-miss cost of double hashing (maphash) plus
bit-array probing exceeds the atomic operations and infrequent mutex acquisition of the
sync.Map miss path. The Bloom filter overhead is a constant factor that dominates the
variable-contention savings at the concurrency levels reachable on a single machine.

**Status:** reproducing
**Supports:** H-001 (falsified), S-001 (fail), S-002 (pass)
**Recorded:** 2026-08-10

---

<!-- Template for further entries:

### E-002 — title

**Claim.**
**Environment:**

```bash
```

**Result:**
**Status:** reproducing | broken
**Supports:**
**Recorded:**

-->

## Benchmark methodology

Filled at P5. Applies to every entry tagged as a benchmark.

- **What is measured:** Mean ns/op (Go benchmark runner) for sync.Map.Load() absent-key
  operations, comparing plain sync.Map vs BloomMap (Bloom filter pre-check enabled).
- **Runs:** 5 per goroutine count, **warm-up:** Go benchmark runner handles warm-up
  automatically.
- **Held constant:** 100,000 pre-populated keys, 100,000 miss keys, 1% Bloom FPR.
  Background writer runs continuously in a separate goroutine.
- **Baseline configuration:** Plain sync.Map with no Bloom filter. Both run on the same
  Go 1.22.2 runtime, same hardware. Same environment, different data structure.
- **Known measurement bias:** Go's benchmark runner reports mean time per operation, not
  p95/p99 latency. The hypothesis was stated in terms of p99 miss latency; the mean is a
  proxy (if the mean is higher, p99 is at least the mean for non-negative distributions).
  Background writer contention is not controlled for — one goroutine's writer produces
  variable dirty-map sizes.
