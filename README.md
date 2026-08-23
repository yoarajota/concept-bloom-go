# Bloom filter pre-check in Go sync.Map: crossover point for miss latency under read-heavy load

> What is the crossover point where a Bloom filter pre-check in Go's sync.Map reduces miss latency under read-heavy concurrent access?

**Concept** `C-002` · **archetype** `implementation`.

Measures whether a Bloom filter pre-check reduces miss latency in Go's sync.Map under
read-heavy concurrent access, by probing the exact combination of goroutine count, miss rate,
and key cardinality at which the Bloom filter's avoidance of the sync.Map internal miss path
pays back its overhead.

## Claim

**H-001** — Under a read-heavy workload where ≥50% of key accesses are to absent keys, a
Bloom filter pre-check on Go's sync.Map achieves at least 20% lower p99 miss latency than a
plain sync.Map Load() at ≥16 goroutines, at the cost of a per-insertion Bloom insertion and
memory for the bit array. → *verdict: **falsified** by [E-003](docs/05-evidence.md#e-003).*

The Bloom filter pre-check was consistently 6–25% **slower** than plain sync.Map across
1–32 goroutines. sync.Map's internal miss path (Go 1.22) is already highly optimised; the
per-miss cost of double hashing plus bit-array probing exceeds the atomic operations and
infrequent mutex acquisition of the sync.Map miss path [E-003](docs/05-evidence.md#e-003).

## Baseline

Compared against **Go sync.Map Load** (Go 1.22) because it is the idiomatic concurrent-map
path in the Go standard library, not a strawman. Reproduce the comparison:
`go test ./bench/... -bench=. -benchmem -count=5 -benchtime=1s` — see
[E-003](docs/05-evidence.md#e-003).

<!-- scorecard:start -->

### Readiness scorecard

_Generated from `.sota/` — do not hand-edit._

| Measure | Value | Meaning |
| :--- | :--- | :--- |
| **TRL** | **4** | Component validated in lab |
| **SRL**  | **3** | High-risk immature technologies identified and prototyped — seams only |
| Composite SRL | 0.370 | standard formulation (all components, diagonal-inclusive) |
| Weakest component | core (0.3333) | lowest component-level SRL |
| Weakest seam | core<->host (IRL 3) | lowest-scoring integration pair |
| Phase | P6 | |
| Hypothesis | falsified | |
| Suitable for | none-yet | |

| Component | Role | TRL | Component SRL |
| :--- | :--- | :-: | :-: |
| `core` | concept | 4 | 0.333 |
| `host` | host | 6 | 0.407 |

| Scenario | Characteristic | Priority | Status |
| :--- | :--- | :--- | :--- |
| S-001 | performance-efficiency | high | fail |
| S-002 | reliability | high | pass |
| S-003 | maintainability | medium | unverified |

<!-- scorecard:end -->

## Try it

```bash
# setup — Go 1.22+ only; everything else is self-contained
go mod tidy

# the headline result in one command
go test ./bench/... -bench=. -benchmem -count=5 -benchtime=1s
```

## How it works

A [Bloom filter](docs/01-theory.md) (bit array with k hash functions) sits in front of
`sync.Map.Load()`. When the filter says "definitely not present," the call returns
immediately without touching sync.Map. When it says "maybe present" (including false
positives), the call falls through to sync.Map. The filter uses `hash/maphash` (Go
standard library) with enhanced double hashing per Kirsch & Mitzenmacher (2006, SRC-004).

Architecture and tradeoffs: [docs/04-tradeoffs.md](docs/04-tradeoffs.md).

## Evidence

| ID | Claim | Command | Result |
| :--- | :--- | :--- | :--- |
| E-001 | Literature survey: 6 sources, 5 full-text | `grep -c SRC docs/01-theory.md` | 6 sources documented |
| E-002 | Bloom filter FPR within 2× analytical prediction | `go test ./bench/... -run TestFPR` | observed 0.0101 vs expected 0.0100 |
| E-003 | Hypothesis falsified: Bloom Map slower at all goroutine counts | `go test ./bench/... -bench=. -count=5` | 6–25% penalty vs plain sync.Map |

Full ledger: [docs/05-evidence.md](docs/05-evidence.md).

## Limitations

- **sync.Map is already fast.** Go 1.22's sync.Map miss path on purely absent keys (keys
  never inserted) involves mainly atomic reads and does not acquire the mutex. The Bloom
  filter overhead exceeds the miss-path cost at all measured concurrency levels.
- **False-positive rate at finite filter sizes.** The enhanced double hashing approximation
  (SRC-004) closes the gap with full independence but does not eliminate it. At m=958k and
  n=100k the observed FPR was within 2× the analytical prediction (E-002).
- **Single-machine measurement only.** The benchmark runs on a 13th Gen Intel i7-13650HX
  laptop. Results on NUMA-scaled servers with more goroutines might differ — but the trend
  (Bloom overhead grows with goroutines) pushes in the wrong direction.
- **Go version sensitivity.** sync.Map internals were replaced in Go 1.24 with a lock-free
  hash trie map (isync.HashTrieMap). The Bloom filter pre-check may behave differently on
  that architecture; this concept targets Go ≤ 1.23.
- **Readiness ceiling: TRL 4.** The component is validated in a lab environment (benchmark on
  a developer laptop). No deployment in an operational environment exists; TRL 7+ is not
  claimable. The SRL of 3 reflects a prototype with gaps (IRL 3 on the core↔host seam).

### What we tried that didn't work

- **FNV-1a double hashing.** Using `hash/fnv` with two salt bytes produced observed FPR 2.6×
  above the analytical prediction (recorded in ADR D-001). Switched to `hash/maphash` with
  two independent seeds, which closed the gap to 1×.
- **No background writer.** An early benchmark without a concurrent writer produced BloomMap
  produced BloomMap with higher per-op latency at every goroutine count — sync.Map's read-only miss path is lock-free and the
  Bloom overhead is pure cost. Adding a background writer forces the dirty map, creating the
  contention the Bloom filter is supposed to avoid, but the overhead still dominates.

## Documents

| Document | Contents |
| :--- | :--- |
| [docs/01-theory.md](docs/01-theory.md) | What the literature establishes, with the `SRC-###` source ledger. |
| [docs/03-log.md](docs/03-log.md) | Append-only working log: what was tried, including what failed. |
| [docs/04-tradeoffs.md](docs/04-tradeoffs.md) | ATAM-lite: drivers, scenarios, sensitivity and tradeoff points, risks. |
| [docs/05-evidence.md](docs/05-evidence.md) | Every claim, its command, environment, and observed result. |
| [docs/adr/](docs/adr/) | Decisions and the options rejected. |
| [.sota/](.sota/) | Machine-readable readiness and quality data. |

## License

MIT
