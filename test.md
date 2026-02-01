C:\Users\99650\stb-demo>go run main.go
STB DEMO (INHIB+PRED+ERROR+FORGET): signals -> blocks -> competition -> prediction -> error-driven learning -> forgetting.
Commands: train | test | reset | board | demo | quit
Suggested demo:
  train  ; reset ; repeat: 1 2 3   (3-5 times)
  test   ; reset ; try:    1 2 4   (watch PRED and ERR + inhibition)
Input tokens separated by spaces. Example: 1 2 1 2 1 2 3 1 2 3 1 2 4
> demo
Demo: investor mode ON, running scripted sequence...
DEMO STEP 1/3: ACCUMULATION -> CRYSTALLIZATION
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
=== BOARD t=012 mode=TRAIN ===
LEARNED: pairs=3 composes=0 actionLinks=6 blocks=15
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[1-2] [2-3] [1-3]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
NOTE: Step 1 reports accumulation and block crystallization (new blocks). STRUCT signals appear in Step 2.
DEMO STEP 2/3: STRUCTURES -> PREDICTION
t=014 INPUT=2  STRUCT=[[1-2]]
           ENERGY_NOW=9.40  SPENT_EP=0.60
+++ PREDICTION UPDATED: [1-2] -> 3 (conf=1.00)
t=015 INPUT=3  STRUCT=[[2-3]]
           ENERGY_NOW=9.40  SPENT_EP=1.20
+++ PREDICTION UPDATED: [2-3] -> 1 (conf=1.00)
t=016 INPUT=1  STRUCT=[[1-3]]
           ENERGY_NOW=9.40  SPENT_EP=1.80
+++ PREDICTION UPDATED: [1-3] -> 2 (conf=1.00)
t=017 INPUT=2  STRUCT=[[1-2]]
           ENERGY_NOW=9.40  SPENT_EP=2.40
t=018 INPUT=3  STRUCT=[[2-3]]
           ENERGY_NOW=9.40  SPENT_EP=3.00
t=019 INPUT=1  STRUCT=[[1-3]]
           ENERGY_NOW=9.40  SPENT_EP=3.60
t=020 INPUT=2  STRUCT=[[1-2] (1>2)]
           ENERGY_NOW=8.80  SPENT_EP=4.80
+++ PREDICTION UPDATED: (1>2) -> 3 (conf=1.00)
t=021 INPUT=3  STRUCT=[[2-3] (2>3)]
           ENERGY_NOW=8.40  SPENT_EP=6.00
+++ PREDICTION UPDATED: (2>3) -> 1 (conf=1.00)
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
DEMO STEP 3/3: MISPREDICTION -> INHIBITION + ERROR-BOOST -> FAST RE-LEARN
+++ AUTO-SENSOR CREATED [4]
t=027 INPUT=4
           CHARGE PAIR [2-4]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (2>4)  mass=0.45/1.00 (+0.45)
           ENERGY_NOW=9.80  SPENT_EP=0.00
t=029 INPUT=2  STRUCT=[[1-2] (1>2)]
           ENERGY_NOW=8.80  SPENT_EP=1.20
+++ PREDICTION UPDATED (error-correct): [1-2] -> 4 (conf=0.73)
+++ PREDICTION UPDATED (error-correct): (1>2) -> 4 (conf=0.91)
+++ LEARNED NEW PAIR BLOCK [2-4]
+++ ATTACHED ACTION ACT_ON_[2-4] <- [2-4]
+++ LEARNED NEW SEQ BLOCK (2>4)
+++ ATTACHED ACTION ACT_ON_(2>4) <- (2>4)
t=030 INPUT=4
           ERROR=[[1-2]:3->4 (1>2):3->4]
           EXPECTATIONS: [1-2]⇒3(conf=1.00) ; (1>2)⇒3(conf=1.00)
           MISPREDICTION: [1-2] expected 3 (conf 1.00) ⇒ got 4 (conf 1.00->0.81)
           MISPREDICTION: (1>2) expected 3 (conf 1.00) ⇒ got 4 (conf 1.00->0.94)
           FIELD ADAPTATION: inhibited=2 | error-boost ttl=6 gain=1.20
           ENERGY_NOW=9.60  SPENT_EP=1.20
=== BOARD t=030 mode=TRAIN ===
LEARNED: pairs=4 composes=1 actionLinks=9 blocks=22
FIELD: energy=9.60/10.00
FIELD: energy_spent_episode=1.20
LAST PAIRS:   [[1-2] [2-3] [1-3] [2-4]]
LAST COMPOSE: [[[1-2]-3]]
EPISODE: structs=[(1>2) [1-2]]
EPISODE: actions=(none)
EPISODE: errors=[(1>2):3⇒4 [1-2]:3⇒4]
FIELD: dominant expectations=[(1>2)⇒4(st=0.94) [1-2]⇒4(st=0.81)]
FIELD: all expectations=[(1>2)⇒4(st=0.94) (2>3)⇒1(st=1.00) [1-2]⇒4(st=0.81) [1-3]⇒2(st=1.00) [2-3]⇒1(st=1.00)]
FIELD: inhib=[(1>2):1.57 [1-2]:0.90 (2>3):0.12]
LEARNING: error-boost=ON ttl=6 gain=1.20
DEMO SUMMARY: learned pairs=4 | composes=1 | actionLinks=9 | blocks=22
> 4 5 6
+++ AUTO-SENSOR CREATED [4]
t=001 INPUT=4
=== BOARD t=001 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=1
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ AUTO-SENSOR CREATED [5]
t=002 INPUT=5
           CHARGE PAIR [4-5]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (4>5)  mass=0.45/1.00 (+0.45)
=== BOARD t=002 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=2
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ AUTO-SENSOR CREATED [6]
t=003 INPUT=6
           CHARGE PAIR [5-6]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (5>6)  mass=0.45/1.00 (+0.45)
=== BOARD t=003 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=3
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 6
t=004 INPUT=4
=== BOARD t=004 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=3
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=005 INPUT=5
           CHARGE PAIR [4-5]  mass=0.80/1.00 (+0.40)
           CHARGE SEQ  (4>5)  mass=0.90/1.00 (+0.45)
           NEAR-CRYSTAL (4>5)
=== BOARD t=005 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=3
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=006 INPUT=6
           CHARGE PAIR [5-6]  mass=0.80/1.00 (+0.40)
           CHARGE SEQ  (5>6)  mass=0.90/1.00 (+0.45)
           NEAR-CRYSTAL (5>6)
=== BOARD t=006 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=3
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 6
t=007 INPUT=4
=== BOARD t=007 mode=TRAIN ===
LEARNED: pairs=0 composes=0 actionLinks=0 blocks=3
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ LEARNED NEW PAIR BLOCK [4-5]
+++ ATTACHED ACTION ACT_ON_[4-5] <- [4-5]
+++ LEARNED NEW SEQ BLOCK (4>5)
+++ ATTACHED ACTION ACT_ON_(4>5) <- (4>5)
t=008 INPUT=5
=== BOARD t=008 mode=TRAIN ===
LEARNED: pairs=1 composes=0 actionLinks=2 blocks=7
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ LEARNED NEW PAIR BLOCK [5-6]
+++ ATTACHED ACTION ACT_ON_[5-6] <- [5-6]
+++ LEARNED NEW SEQ BLOCK (5>6)
+++ ATTACHED ACTION ACT_ON_(5>6) <- (5>6)
t=009 INPUT=6
=== BOARD t=009 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 6
t=010 INPUT=4
=== BOARD t=010 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=011 INPUT=5
=== BOARD t=011 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=012 INPUT=6
=== BOARD t=012 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 6
t=013 INPUT=4
=== BOARD t=013 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=014 INPUT=5
=== BOARD t=014 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=015 INPUT=6
=== BOARD t=015 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 6
t=016 INPUT=4
=== BOARD t=016 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=017 INPUT=5  STRUCT=[[4-5]]
           ENERGY_NOW=9.40  SPENT_EP=0.60
=== BOARD t=017 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=9.40/10.00
FIELD: energy_spent_episode=0.60
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=[[4-5]]
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=(none)
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ PREDICTION UPDATED: [4-5] -> 6 (conf=1.00)
t=018 INPUT=6  STRUCT=[[5-6]]
           CHARGE COMP [[4-5]-6]  mass=0.28/1.00 (+0.28)
           CHARGE PRED [4-5] -> 6  w=0.65 (+0.65)
           ENERGY_NOW=9.40  SPENT_EP=1.20
=== BOARD t=018 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=9.40/10.00
FIELD: energy_spent_episode=1.20
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=[[4-5] [5-6]]
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=[[4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 9
t=019 INPUT=4
=== BOARD t=019 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=[[4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=020 INPUT=5  STRUCT=[(4>5)]
           ENERGY_NOW=9.40  SPENT_EP=0.60
=== BOARD t=020 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=11
FIELD: energy=9.40/10.00
FIELD: energy_spent_episode=0.60
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=[(4>5)]
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=[[4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ AUTO-SENSOR CREATED [9]
+++ PREDICTION UPDATED: (4>5) -> 9 (conf=1.00)
t=021 INPUT=9
           CHARGE PAIR [5-9]  mass=0.40/1.00 (+0.40)
           CHARGE SEQ  (5>9)  mass=0.45/1.00 (+0.45)
           CHARGE PRED (4>5) -> 9  w=0.65 (+0.65)
=== BOARD t=021 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=12
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.60
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=[(4>5)]
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=[(4>5)⇒9(st=1.00)]
FIELD: all expectations=[(4>5)⇒9(st=1.00) [4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
> 4 5 9
t=022 INPUT=4
=== BOARD t=022 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=12
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.00
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=(none)
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=(none)
FIELD: all expectations=[(4>5)⇒9(st=1.00) [4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
t=023 INPUT=5  STRUCT=[[4-5]]
           ENERGY_NOW=9.40  SPENT_EP=0.60
=== BOARD t=023 mode=TRAIN ===
LEARNED: pairs=2 composes=0 actionLinks=4 blocks=12
FIELD: energy=9.40/10.00
FIELD: energy_spent_episode=0.60
LAST PAIRS:   [[4-5] [5-6]]
EPISODE: structs=[[4-5]]
EPISODE: actions=(none)
EPISODE: errors=(none)
FIELD: dominant expectations=[[4-5]⇒6(st=1.00)]
FIELD: all expectations=[(4>5)⇒9(st=1.00) [4-5]⇒6(st=1.00)]
FIELD: inhib=(none)
LEARNING: error-boost=OFF
+++ PREDICTION UPDATED (error-correct): [4-5] -> 9 (conf=0.91)
+++ LEARNED NEW PAIR BLOCK [5-9]
+++ ATTACHED ACTION ACT_ON_[5-9] <- [5-9]
+++ LEARNED NEW SEQ BLOCK (5>9)
+++ ATTACHED ACTION ACT_ON_(5>9) <- (5>9)
t=024 INPUT=9
           CHARGE PAIR [5-9]  mass=0.60/1.00 (+0.20)
           CHARGE SEQ  (5>9)  mass=0.60/1.00 (+0.15)
           CHARGE COMP [[4-5]-9]  mass=0.62/1.00 (+0.62)
           CHARGE PRED [4-5] -> 9  w=3.85 (+3.85)
           ERROR=[[4-5]:6->9]
           MISPREDICTION: [4-5] expected 6 (conf 1.00) ⇒ got 9 (conf 1.00->0.94)
           ENERGY_NOW=10.00  SPENT_EP=0.60
=== BOARD t=024 mode=TRAIN ===
LEARNED: pairs=3 composes=0 actionLinks=6 blocks=16
FIELD: energy=10.00/10.00
FIELD: energy_spent_episode=0.60
LAST PAIRS:   [[4-5] [5-6] [5-9]]
EPISODE: structs=[[4-5]]
EPISODE: actions=(none)
EPISODE: errors=[[4-5]:6⇒9]
FIELD: dominant expectations=[[4-5]⇒9(st=0.94)]
FIELD: all expectations=[(4>5)⇒9(st=1.00) [4-5]⇒9(st=0.94)]
FIELD: inhib=[[4-5]:0.90]
LEARNING: error-boost=ON ttl=6 gain=1.20