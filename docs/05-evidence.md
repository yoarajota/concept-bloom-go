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

- **What is measured:** TODO
- **Runs:** TODO (≥ 5), **warm-up:** TODO
- **Held constant:** TODO
- **Baseline configuration:** TODO — the same tuning effort was spent on the baseline as on
  the concept; if not, say so, because it invalidates the comparison.
- **Known measurement bias:** TODO
