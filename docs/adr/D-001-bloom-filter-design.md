# D-001 — Bloom filter design: double hashing with maphash

**Date:** 2026-08-10
**Status:** accepted

## Context

The Bloom filter pre-check for sync.Map needs a hash function strategy. The filter must be fast enough that the pre-check cost is lower than the sync.Map miss-path it avoids. Kirsch & Mitzenmacher (2006, SRC-004) show that two hash functions simulate k independent hash functions via enhanced double hashing with asymptotically equivalent false-positive rates.

## Decision

Use Go's `hash/maphash` with two fixed seeds for the two base hash functions, applying enhanced double hashing `g_i(x) = (h1(x) + i·h2(x) + i²) mod m`. `maphash` is the fastest hash in Go's standard library, and the i² term reduces the gap between the two-hash approximation and k truly independent hash functions at finite m (SRC-004).

## Rejected options

| Option | Reason |
| :--- | :--- |
| FNV-1a from `hash/fnv` | Two FNV hashes on the same key with a salt byte are too correlated; observed FPR was 2.6x the analytical prediction |
| `crypto` hashes (SHA, Blake) | Cryptographic hashes have higher per-invocation overhead; the Bloom filter is on the hot path and cryptographic security is irrelevant |
| External dependency (xxhash, murmur3) | Adds a module dependency with no gain over maphash — maphash is already in the standard library and is designed for hash-table workloads |
| k independent hash functions | Requires k·Hash() calls instead of 2; the Kirsch-Mitzenmacher result avoids this with no cost in false-positive rate |

## Consequences

- **No external dependencies.** The concept remains a single-module Go project using only the standard library. This satisfies R5 (complexity budget) and avoids the complexity-ledger scrutiny for external deps.
- **Deterministic but not reproducible across builds.** `maphash.MakeSeed()` generates per-process seeds, so the same key produces different bit-array patterns across runs. Filter correctness (no false negatives) is unaffected. Benchmark variance across runs is the expected tradeoff.
- **False-positive rate at finite m may slightly exceed the formula.** Enhanced double hashing closes the gap but does not eliminate it entirely. The conformance test (E-002) tolerates up to 2x the analytical prediction.
