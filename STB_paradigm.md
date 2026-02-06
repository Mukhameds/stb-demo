
---


# STB (Signal Theory of Being)
## Signal-reactive computational paradigm


---

## 1. Scope of definition

STB is a computational paradigm in which:

* the basic unit of computation is a **signal**;
* execution logic is distributed across **reactive blocks**;
* system behavior is determined by the **dynamics of signals and structures**, not by an algorithm.

STB is applicable to systems where:

* computation is continuous in time;
* the system structure can change during execution;
* there is no global controlling algorithm.

---

## 2. Core entities

### 2.1 Signal

A **signal** is an atomic discrete event described by the tuple:

```
Signal = { type, value, mass, time, origin }
```

Minimum requirements:

* a signal has a type (category);
* a signal may carry numerical or structural payload;
* a signal propagates through the system and may generate other signals.

A signal is not a function call and does not imply a response.

---

### 2.2 Block

A **block** is an autonomous reactive element possessing:

* local state;
* a set of reaction rules;
* the ability to generate signals.

Formally, a block is a mapping:

```
(BlockState, Signal) → { BlockState', Signal* }
```

A block:

* does not know the global system state;
* does not directly control other blocks;
* reacts only to incoming signals and its own state.

---

### 2.3 Reaction

A **reaction** is a local computation inside a block that may:

* change the block’s state;
* generate zero or more new signals;
* have no external effect.

A reaction is not a transaction and does not guarantee a deterministic system-level outcome.

---

## 3. Computational cycle

The minimal STB cycle:

```
Signal → Block → Reaction → Signal*
```

The system operates as a **field of interacting reactions**.

Global execution order:

* is not defined;
* may be asynchronous;
* may be multithreaded;
* does not affect the validity of the paradigm.

---

## 4. Absence of a global algorithm

In STB:

* there is no main function;
* there is no central logic scheduler;
* there is no fixed execution scenario.

System behavior is the **result of competing block reactions**.

---

## 5. Internal signals and autonomous activity

STB allows signal generation without external input.

Sources of internal signals include:

* internal timers;
* accumulated states;
* threshold conditions;
* structural conflicts;
* degradation or reinforcement of blocks.

This allows the system to:

* reprocess existing structures;
* change priorities;
* perform internal activity cycles.

---

## 6. Structural memory

Memory in STB is implemented through the **stability of blocks and their connections**.

Invariants:

* frequently activated blocks are reinforced;
* rarely used blocks degrade;
* unused blocks are removed.

No separate memory storage mechanism is required.

---

## 7. Learning and adaptation

STB does not separate:

* learning;
* execution;
* adaptation.

System change occurs through:

* changes in block state;
* creation of new blocks;
* removal of blocks;
* modification of reaction rules.

All changes are local.

---

## 8. Block generation

Blocks may generate other blocks when specified conditions are met.

Generation conditions may include:

* accumulation of signal mass;
* repetition of patterns;
* block competition;
* prediction errors.

Block generation is not a global operation.

---

## 9. Block types (generalized)

STB does not fix block types, but allows:

* reactive;
* accumulating;
* inertial;
* competitive;
* inhibitory;
* structure-forming;
* generative blocks.

A block type is determined by its reaction rules.

---

## 10. Determinism

STB allows:

* deterministic block reactions;
* stochastic elements inside blocks.

System-level determinism is neither guaranteed nor required.

---

## 11. Scaling

Scaling of an STB system is achieved through:

* increasing the number of blocks;
* increasing signal density;
* increasing structural depth.

Adding blocks does not require modifying existing blocks.

---

## 12. Comparison with classical paradigms

**Classical → STB**

* Function → Reaction
* Algorithm → Dynamics
* Global control → Local rules
* Data → Structures
* Memory as storage → Memory as stability

---

## 13. Paradigm limitations

STB does not define:

* implementation language;
* concurrency model;
* signal format;
* scheduling strategy.

STB defines **principles of interaction between computational elements**.

---

## 14. Formal statement

STB describes computational systems in which:

> computation is a continuous process of reactions,
> and system logic is formed and modified
> through the dynamics of signals and structures.

---

