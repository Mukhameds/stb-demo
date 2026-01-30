C:\Users\99650\stb-demo>go run main.go
STB DEMO (INHIB+PRED+ERROR+FORGET): signals -> blocks -> competition -> prediction -> error-driven learning -> forgetting.
Commands: train | test | reset | board | quit
Suggested demo:
  train  ; reset ; repeat: 1 2 3   (3-5 times)
  test   ; reset ; try:    1 2 4   (watch PRED and ERR + inhibition)
Input tokens separated by spaces. Example: 1 2 1 2 1 2 3 1 2 3 1 2 4
> demo
Demo: investor mode ON, running scripted sequence...
PHASE 1/3: ACCUMULATION -> CRYSTALLIZATION
+++ AUTO-SENSOR CREATED [1]
+++ AUTO-SENSOR CREATED [2]
t=002 INPUT=2
           CHARGE PAIR [1-2]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (1>2)  mass=0.45/1.00 (+0.45)
           ENERGY_NOW=10.00  SPENT_EP=0.00
+++ AUTO-SENSOR CREATED [3]
t=003 INPUT=3
           CHARGE PAIR [2-3]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (2>3)  mass=0.45/1.00 (+0.45)
           ENERGY_NOW=10.00  SPENT_EP=0.00
t=004 INPUT=1
           CHARGE PAIR [1-3]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (3>1)  mass=0.45/1.00 (+0.45)
           ENERGY_NOW=10.00  SPENT_EP=0.00
t=005 INPUT=2
           CHARGE PAIR [1-2]  mass=0.80/1.00 (+0.40)
           CHARGE SEQ  (1>2)  mass=0.90/1.00 (+0.45)
           NEAR-CRYSTAL (1>2)
           ENERGY_NOW=10.00  SPENT_EP=0.00
t=006 INPUT=3
           CHARGE PAIR [2-3]  mass=0.80/1.00 (+0.40)
           CHARGE SEQ  (2>3)  mass=0.90/1.00 (+0.45)
           NEAR-CRYSTAL (2>3)
           ENERGY_NOW=10.00  SPENT_EP=0.00
t=007 INPUT=1
           CHARGE PAIR [1-3]  mass=0.80/1.00 (+0.40)
           CHARGE SEQ  (3>1)  mass=0.90/1.00 (+0.45)
           NEAR-CRYSTAL (3>1)
           ENERGY_NOW=10.00  SPENT_EP=0.00
+++ LEARNED NEW PAIR BLOCK [1-2]
+++ ATTACHED ACTION ACT_ON_[1-2] <- [1-2]
+++ LEARNED NEW SEQ BLOCK (1>2)
+++ ATTACHED ACTION ACT_ON_(1>2) <- (1>2)
+++ LEARNED NEW PAIR BLOCK [2-3]
+++ ATTACHED ACTION ACT_ON_[2-3] <- [2-3]
+++ LEARNED NEW SEQ BLOCK (2>3)
+++ ATTACHED ACTION ACT_ON_(2>3) <- (2>3)
+++ LEARNED NEW PAIR BLOCK [1-3]
+++ ATTACHED ACTION ACT_ON_[1-3] <- [1-3]
+++ LEARNED NEW SEQ BLOCK (3>1)
+++ ATTACHED ACTION ACT_ON_(3>1) <- (3>1)
PHASE 2/3: STRUCTURES -> PREDICTIONS
t=014 INPUT=2  STRUCT=[[1-2]]
           ENERGY_NOW=9.40  SPENT_EP=0.60
+++ EMERGENT EXPECTATION: [1-2] ⇒ 3 (stability=1.00)
t=015 INPUT=3  STRUCT=[[2-3]]
           ENERGY_NOW=9.40  SPENT_EP=1.20
+++ EMERGENT EXPECTATION: [2-3] ⇒ 1 (stability=1.00)
t=016 INPUT=1  STRUCT=[[1-3]]
           ENERGY_NOW=9.40  SPENT_EP=1.80
+++ EMERGENT EXPECTATION: [1-3] ⇒ 2 (stability=1.00)
t=017 INPUT=2  STRUCT=[[1-2]]
           ENERGY_NOW=9.40  SPENT_EP=2.40
t=018 INPUT=3  STRUCT=[[2-3]]
           ENERGY_NOW=9.40  SPENT_EP=3.00
t=019 INPUT=1  STRUCT=[[1-3]]
           ENERGY_NOW=9.40  SPENT_EP=3.60
t=020 INPUT=2  STRUCT=[[1-2] (1>2)]
           ENERGY_NOW=8.80  SPENT_EP=4.80
+++ EMERGENT EXPECTATION: (1>2) ⇒ 3 (stability=1.00)
t=021 INPUT=3  STRUCT=[[2-3] (2>3)]
           ENERGY_NOW=8.40  SPENT_EP=6.00
+++ EMERGENT EXPECTATION: (2>3) ⇒ 1 (stability=1.00)
t=022 INPUT=1  STRUCT=[[1-3]]
           ENERGY_NOW=8.60  SPENT_EP=6.60
t=023 INPUT=2  STRUCT=[[1-2]]
           ENERGY_NOW=8.00  SPENT_EP=8.00
+++ LEARNED NEW COMPOSE BLOCK [[1-2]-3]
+++ ATTACHED ACTION ACT_ON_[[1-2]-3] <- [[1-2]-3]
t=024 INPUT=3  STRUCT=[[2-3]]
           ENERGY_NOW=7.40  SPENT_EP=9.40
=== BOARD t=024 mode=TRAIN ===
LEARNED: pairs=3 composes=1 actionLinks=7 blocks=17
FIELD: energy=7.40/10.00
FIELD: energy_spent_episode=9.40
LAST PAIRS:   [[1-2] [2-3] [1-3]]
LAST COMPOSE: [[[1-2]-3]]
EPISODE: structs=[(1>2) (2>3) [1-2] [1-3] [2-3]]
EPISODE: actions=[ACT_ON_[1-2] ACT_ON_[2-3]]
EPISODE: errors=(none)
FIELD: dominant expectations=[(1>2)⇒3(st=1.00) (2>3)⇒1(st=1.00) [1-2]⇒3(st=1.00) [1-3]⇒2(st=1.00) [2-3]⇒1(st=1.00)]
FIELD: inhib=[(2>3):0.39 (1>2):0.32]
LEARNING: error-boost=OFF
PHASE 3/3: EXPECTATION COLLAPSE -> INHIBITION (ERROR-CONTEXT)
+++ NEW TOKEN REGISTERED [4] (no learning)
t=029 INPUT=2  STRUCT=[[1-2] (1>2)]
           ENERGY_NOW=8.80  SPENT_EP=1.20
t=030 INPUT=4
           COLLAPSE=[(1>2):3⇒4 [1-2]:3⇒4]
           EMERGENT EXPECTATIONS: (1>2)⇒3(stability=1.00) ; [1-2]⇒3(stability=1.00)
           EXPECTATION COLLAPSE: (1>2) expected 3 (stability 1.00) ⇒ got 4
           EXPECTATION COLLAPSE: [1-2] expected 3 (stability 1.00) ⇒ got 4
           FIELD RESPONSE: inhibition=2 hypotheses | error-context ttl=6 gain=1.20
           ENERGY_NOW=9.60  SPENT_EP=1.20
=== BOARD t=030 mode=TEST ===
LEARNED: pairs=3 composes=1 actionLinks=7 blocks=18
FIELD: energy=9.60/10.00
FIELD: energy_spent_episode=1.20
LAST PAIRS:   [[1-2] [2-3] [1-3]]
LAST COMPOSE: [[[1-2]-3]]
EPISODE: structs=[(1>2) [1-2]]
EPISODE: actions=(none)
EPISODE: errors=[(1>2):3⇒4 [1-2]:3⇒4]
FIELD: dominant expectations=[(1>2)⇒3(st=1.00) [1-2]⇒3(st=1.00)]
FIELD: all expectations=[(1>2)⇒3(st=1.00) (2>3)⇒1(st=1.00) [1-2]⇒3(st=1.00) [1-3]⇒2(st=1.00) [2-3]⇒1(st=1.00)]
FIELD: inhib=[(1>2):1.57 [1-2]:0.90 (2>3):0.12]
LEARNING: error-boost=ON ttl=6 gain=1.20
DEMO SUMMARY: learned pairs=3 | composes=1 | actionLinks=7 | blocks=18