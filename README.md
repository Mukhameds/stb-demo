
---

# STB-DEMO

Minimal deterministic prototype of a signal-reactive adaptive architecture.

---

## What This Repository Is

STB-DEMO is a compact engineering prototype that demonstrates:

* online structural formation from repeated signals,
* prediction derived from structure,
* local error-driven adaptation,
* bounded growth with controlled pruning,
* deterministic and inspectable behavior.

This repository is not a task-optimized AI model.
It is a systems-level architecture validation.

---

## Execution Model

Each tick processes signals through local reactions:

```
Signal → Block.React() → new Signals → competition → winner → next tick
```

There is:

* no global controller,
* no centralized policy function,
* no separate training phase.

All behavior emerges from signal propagation between independent blocks.

---

## Core Concepts

### Signal

A discrete event carrying:

* type (input, structure, prediction, error, action),
* value,
* activation mass.

Signals are the only way components interact.

---

### Block

An autonomous processing unit that:

* reacts to incoming signals,
* may emit new signals,
* maintains only local state.

Blocks do not coordinate directly.
They influence each other only through the shared signal field.

---

## Implemented Mechanisms

### 1. Structural Formation

Repeated input patterns accumulate internal mass.

When a threshold is reached, new structural blocks are created:

* Pair blocks: `[A-B]`
* Sequence blocks: `(A>B)`
* Compose blocks: `[[A-B]-C]`

This enables incremental structural growth without dataset training.

---

### 2. Structural Prediction

Structures maintain transition statistics.

When activated strongly enough, a structure arms an expectation:

```
[A-B] ⇒ C
```

Prediction is produced by structural dominance, not by a separate inference pass.

---

### 3. Error-Driven Adaptation

If actual input contradicts an armed expectation:

* an error signal is emitted,
* incorrect transition weights are dampened,
* alternative transitions are reinforced,
* temporary amplification accelerates adaptation.

No retraining loop is executed.
Adaptation occurs online.

---

### 4. Competition and Inhibition

Activated structures accumulate mass.

The strongest structure:

* becomes the winner for the tick,
* suppresses alternatives through inhibition.

This stabilizes behavior without centralized selection logic.

---

### 5. Energy Constraint

Actions consume energy.

Low energy reduces signal strength.

This introduces a simple resource constraint into activation dynamics.

---

### 6. Controlled Forgetting

Inactive learned blocks are pruned after a configurable window.

Structures are preserved if they:

* remain predictive,
* contain significant transition mass,
* were recently active.

This prevents unbounded structural growth.

---

## Demonstration Scenario

Run the interactive demo:

```
go run main.go
```

Then execute:

```
demo
```

The demo performs three stages:

### Stage 1 — Structural Formation

Repeated sequences:

```
1 2 1 2 1 2
2 3 2 3 2 3
```

Result:

* Pair structures `[1-2]` and `[2-3]` are formed.

---

### Stage 2 — Stable Prediction

Repeated pattern:

```
1 2 3
```

Result:

* Structure `[1-2]` consistently predicts `3`.
* Prediction confidence increases.

---

### Stage 3 — Misprediction and Adaptation

Input:

```
1 2 4
```

Result:

* Error is detected.
* Inhibition is applied.
* Prediction shifts from `3` to `4` after repeated evidence.
* No retraining cycle occurs.

All internal events are visible in logs.

---

## What This Prototype Validates

This prototype demonstrates that:

* online structural learning can occur without a training phase,
* prediction can emerge from structure competition,
* adaptation can occur via local error signals,
* structural pruning can bound system growth,
* the entire process can remain deterministic and inspectable.

---

## Scope

This implementation is intentionally minimal.

It does not attempt:

* large-scale performance,
* task benchmarking,
* optimization for throughput.

Its purpose is architectural validation.

---

## Potential Extensions

This architecture could be extended toward:

* distributed signal-based systems,
* adaptive edge agents,
* online continual learning environments,
* resource-constrained adaptive agents.

---

This repository represents a compact, fully inspectable foundation for signal-driven adaptive computation.

---

See STB_paradigm.md for the formal definition of the computational paradigm.
