---
name: sota
description: Work the current phase of this SOTA concept project. Use whenever asked what to do next, to advance a phase, to record evidence, to score readiness, or before claiming any gate passes. Triggers: /sota, what's next, advance the phase, run the gates, score readiness, publish this concept.
---

# Work this concept project

This repository is a SOTA concept project (`C-002`, archetype `implementation`). It
answers one question: What is the crossover point where a Bloom filter pre-check in Go's sync.Map reduces miss latency under read-heavy concurrent access?

## Step 1 — Ask the tool, don't guess

```bash
FW=${SOTA_FRAMEWORK:-../sota-theoretical-framework}
python3 $FW/tools/sota.py next .
```

If that path is wrong, find the framework or clone it:

```bash
git clone https://github.com/yoarajota/sota-theoretical-framework ../sota-theoretical-framework
```

`next` prints the current phase, the failing gate, the **exact files to read for this phase**,
and the commands to run when the work is done. It is the routing layer — do not reconstruct it
by reading the whole framework.

## Step 2 — Read only what it named

Read the workflow file and the instruments `next` listed. Do not load the other phases'
material; it is not relevant and costs context you need for the work.

Always also load, once per session:
- `$FW/framework/20-rules.md` — binding, and overrides convenience.
- `$FW/framework/40-limitations.md` — before reporting any number, so you know which rules the
  tooling enforces and which rest on your honesty.

## Step 3 — Do the phase work

Follow the workflow's steps in order. Do not skip ahead to a later phase: each phase's output is
the next one's input.

The rules broken most often, in order:

1. **Unevidenced adjectives.** "Efficient", "fast", "robust" with no `E-###`. Delete the word or
   produce the measurement.
2. **Inflated readiness.** TRL 7 requires a real operational environment. A load generator on
   your machine is TRL 6 at best.
3. **Recalled citations.** Never cite a source you did not fetch in this session. If a source
   resists, escalate before giving up — `WebSearch` → `WebFetch` → cache the PDF and extract it
   locally with `pypdf` in a venv → the **`playwright-cli` skill** (delegated to a sub-agent) for
   pages a plain fetch cannot render → open-access mirrors (arXiv, OSTI, DTIC, NTRS, university
   copies, the author's homepage). Full ladder:
   `$FW/workflows/03-theory-pass.md § The sourcing toolkit`. Never bypass a paywall — record
   `Access: abstract-only` instead.
4. **Host installs.** Never `apt`/`brew`/`npm -g` anything as part of setup. Supporting
   services run as version-pinned containers declared in the repo; language tooling stays
   project-local (R10). A reader must reproduce your environment on a clean machine and remove
   it afterwards.
5. **Unpaid complexity.** Anything structural with no scenario requiring it gets deleted, not
   justified after the fact.
6. **Hypothesis drift.** Never edit the claim to match a disappointing measurement. A falsified
   hypothesis is a publishable result.

## Step 4 — Close the loop

```bash
python3 $FW/tools/sota.py validate  .                # must exit 0
python3 $FW/tools/sota.py srl       . --write        # if components/TRL/IRL changed
python3 $FW/tools/sota.py scorecard . --write
python3 $FW/tools/sota.py next      .                # confirm what moved
```

Before claiming evidence reproduces, prove it:

```bash
python3 $FW/tools/sota.py validate . --run-evidence  # actually executes every E-### command
```

## Step 5 — Advance without asking

**When the gate passes, set the next `phase:` and keep going.** The gate is the checkpoint and
it just passed; asking a human to confirm what `validate` verified adds nothing and turns a
seven-phase lifecycle into seven interruptions.

Stop and ask only when one of these is true, and say which:

- a gate fails on something only a human can decide (a falsified hypothesis, a mis-tuned
  baseline — not "the tests fail", which you fix);
- a genuine fork: probe exit versus continuing, or a result that changes what the concept *is*;
- an outward-facing or costly action: creating a public repo, pushing, paid infrastructure;
- the human asked for a specific scope and it is complete;
- proceeding would require breaking a rule.

Otherwise: report the milestone in one line and continue.

## Step 6 — Report honestly

State what changed, which gate moved, and **whether any readiness level went down**. A dropped
level is a normal outcome of honest scoring. Never hide one, and never report a gate as passing
without having seen `validate` exit 0 in this session.

## Stopping is allowed

If the question is answered and hardening it would not change the answer, publish as a probe
rather than abandoning: set `status: paused`, list skipped gates in `not_applicable`, accept the
TRL 3 cap, and complete P6. See `$FW/framework/10-lifecycle.md § Publishing early`.
