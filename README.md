


# STB Demo — stb-demo-v1

This repository contains a minimal, runnable demonstration of the STB (Signal–Block) paradigm.

It is **not a framework**, not a benchmark, and not a machine learning model.  
It exists to show one thing: how adaptive behavior can emerge from signals, competition, and inhibition — without datasets, gradients, or reward functions.

The code is intentionally small and self-contained.  
Think of it as a reference artifact, not a product.

---

## What you will see

There are:

- no neural networks  
- no training datasets  
- no backpropagation  
- no explicit reward optimization  

Instead, the system exhibits:

- accumulation of signal mass over time  
- crystallization of structures from repeated exposure  
- competition and dominance between structures  
- emergent expectations (not hard-coded rules)  
- collapse of expectations when novelty appears  
- field-level inhibition and contextual adaptation  
- gradual forgetting of unused structures  

All behavior arises from local interactions.  
Nothing is optimized end-to-end.

---

## Running the demo

Requirements:
- Go 1.20+
````
go run main.go
````

Inside the program, run:

```
demo
```

This triggers a scripted, three-phase scenario intended for inspection and discussion.

---

## Demo phases

**Phase 1 — Accumulation → Crystallization**
Repeated signals build mass and stabilize into structures.

**Phase 2 — Structures → Expectations**
Dominant structures begin to generate expectations about what comes next.

**Phase 3 — Novelty → Collapse → Inhibition**
A new signal violates expectations, causing collapse and a temporary inhibitory field response.

You can inspect internal state at any time with:

```
board
```

---

## Why this is not classical ML

* There is no training vs inference split
* No explicit loss or objective
* No probabilistic model over data
* No gradient-based optimization

Expectations emerge through competition and dominance, not supervision.

---

## Why this exists

Most adaptive systems today are built around parameter fitting.
This demo explores a different path: adaptation through structure formation and field dynamics.

It is a small proof-of-mechanism that points toward:

* decentralized agents
* long-lived adaptive systems
* local, low-compute intelligence (ARA concept)

---

## Status

* Frozen reference version: **stb-demo-v1**
* Purpose: demonstration and discussion
* Scaling and architecture work are intentionally out of scope here

```
