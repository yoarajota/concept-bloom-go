# Agent contract — Bloom filter pre-check in Go sync.Map: crossover point for miss latency under read-heavy load

This repository is a **SOTA concept project** (`C-002`, archetype `implementation`).
It answers exactly one question:

> What is the crossover point where a Bloom filter pre-check in Go's sync.Map reduces miss latency under read-heavy concurrent access?

## Before doing anything

Use the `/sota` skill — it wraps this. Manually:

```bash
FW=${SOTA_FRAMEWORK:-../sota-theoretical-framework}
python3 $FW/tools/sota.py next .
```

`next` prints the current phase, the first failing gate, the **exact files to read for this
phase**, and the commands to run when the work is done. Read only what it names — loading the
whole framework wastes the context you need for the work. Do not skip ahead to a later phase.

If the framework repository is not available locally, clone it — this project's gates cannot be
evaluated without it:
`git clone https://github.com/yoarajota/sota-theoretical-framework ../sota-theoretical-framework`

## Rules that bind you here

The full set is `framework/20-rules.md`. The ones violated most often:

- **No unevidenced claim.** Every capability, number, or comparison in `README.md` or `docs/`
  carries an `E-###` that resolves to an entry with a reproduction command, an environment, and
  an observed result. Adjectives without an ID get deleted, not softened.
- **No hand-written scores.** `composite_srl`, `translated_srl`, and component SRL values are
  written only by `sota.py srl --write`.
- **Readiness is descriptive.** Claim the highest level whose evidence reproduces *today*. If
  something broke, lower the level and say so.
- **Baseline or nothing.** "Faster" requires the named baseline, the metric, the conditions, and
  the `E-###`.
- **Complexity must be purchased.** Anything structural traces to a scenario `S-###` in
  `.sota/quality-gates.yaml` and is recorded in `complexity_ledger`. Otherwise remove it.
- **One question.** A second interesting idea goes to the framework's `registry/backlog.md`,
  not into this repository.
- **Real citations only.** Sources must have been fetched and read, not recalled.

## Definition of done for any change

```bash
python3 <framework>/tools/sota.py validate .            # gates + rules, must exit 0
python3 <framework>/tools/sota.py srl . --write         # if components/TRL/IRL changed
python3 <framework>/tools/sota.py scorecard . --write   # refresh the README block
<the iso5055 runner from .sota/quality-gates.yaml>      # must exit 0
```

**Advance phases on your own.** When `validate` exits 0 for the current phase, set the next
`phase:` and continue — do not ask permission. The gate is the checkpoint. Stop only for a
failed gate needing a human decision, a genuine fork, an outward-facing action, or scope the
human limited (`framework/10-lifecycle.md § Advancing between phases`).

**Before you run low on context**, append an `L-###` entry to `docs/03-log.md` with
`Disposition: open` describing what is half-finished. `sota.py next` prints it to the next
session. This is the only state that survives a session boundary.

Report: what changed, which gate state moved, and whether any readiness level went **down**.
A dropped level is a normal outcome and is stated plainly, never hidden.

## Commit messages

History is part of the public record — it is read like the README. Commits follow
[Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

- `feat:` a new capability in the public surface; `fix:` a bug fix; `docs:` README,
  theory, evidence entries, ADRs, and the artifact narrative/page; `perf:` benchmarks
  and anything touching measured performance; `test:` the test suite; `refactor:`
  behaviour-preserving restructuring; `build:` build tooling; `ci:` CI configuration;
  `chore:` everything else. Other types may be used when they fit better.
- A scope is optional and names the part: `feat(core):`, `fix(pagination):`.
- Breaking changes to the public surface get `!` after type/scope, or a
  `BREAKING CHANGE:` footer.
- Description in imperative mood, lowercase, no trailing period:
  `fix: guard page number against empty cursor`.

Overlay rules for this repository:

- Cite the stable ID when the change has one — the message points at the record,
  it does not re-narrate it: `docs: add E-004 depth sweep`, `test: cover E-003
  failure modes`, `docs: D-001 query building decision`.
- One concern per commit. A commit that fixes a bug *and* rewrites the README is two commits.
- No mentions of assistants, AI, or tooling — the history is the project's own.
- No process-state messages ("finished phase X") — commit content, not progress.
- Never rewrite published history to hide a reversal; the history itself shows it.

## Stable IDs used in this repository

| Prefix | Meaning | Lives in |
| :--- | :--- | :--- |
| `C-###` | Concept | `.sota/concept.yaml` |
| `H-###` | Hypothesis | `.sota/concept.yaml` |
| `E-###` | Evidence | `docs/05-evidence.md` |
| `L-###` | Working-log entry | `docs/03-log.md` (append-only) |
| `SRC-###` | Source, with an `Access:` level | `docs/01-theory.md` |
| `S-###` | Quality-attribute scenario | `.sota/quality-gates.yaml` |
| `D-###` | Architecture decision record | `docs/adr/D-###-*.md` |
| `SP-###` / `TP-###` | Sensitivity / tradeoff point | `docs/04-tradeoffs.md` |
| `R-###` / `NR-###` | Risk / non-risk | `docs/04-tradeoffs.md` |
| `X-###` | ISO 5055 exemption | `.sota/quality-gates.yaml` |

IDs are never reused or renumbered. Retired IDs stay in place marked `retired`.
