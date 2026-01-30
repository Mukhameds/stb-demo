package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

var showPredEvents = true

func clearBoolMap(m map[string]bool) {
	for k := range m {
		delete(m, k)
	}
}

func clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

func clearFloatMap(m map[string]float64) {
	for k := range m {
		delete(m, k)
	}
}

type Kind string

const (
	K_SENS   Kind = "SENS"   // sensory input token
	K_ACT    Kind = "ACT"    // activation signal from a block
	K_STRUCT Kind = "STRUCT" // learned structure signal

	K_PRED  Kind = "PRED" // prediction: expects next sensory token
	K_ERR   Kind = "ERR"  // prediction error (mismatch)
	K_NOTE  Kind = "NOTE"
	K_INHIB Kind = "INHIB" // inhibition (competition / suppression)

	K_DRIVE  Kind = "DRIVE"  // (reserved) drive towards action (later)
	K_ACTION Kind = "ACTION" // action output (visible)
)

type Signal struct {
	Kind  Kind
	Value string
	Mass  float64
	Time  int
	From  string
}

// ---- Block interface (pure STB: reacts to signals and emits signals) ----

type Block interface {
	ID() string
	React(s Signal, ctx *Context) []Signal
	Tick(ctx *Context) []Signal
}

// ---- Context holds global field state (blocks remain local laws) ----

type Context struct {
	Tick         int
	RecentActs   []Signal
	RecentStruct []Signal

	Blocks map[string]Block
	Order  []string

	// auto-seeded sensors
	Sensors map[string]bool

	SeenPairs    map[string]float64
	SeenComposes map[string]float64

	SeenSeq map[string]float64 // "1>2" -> learn mass

	// --- adjacency memory (clean pair learning) ---
	PrevSens string
	LastSens string

	PrevStructSet map[string]bool // structures seen on previous tick
	ThisStructSet map[string]bool // structures seen on current tick

	// --- modes ---
	LearningEnabled bool // TRAIN=true, TEST=false

	// --- field resource budget (energy) ---
	Energy      float64 // current energy
	EnergyMax   float64 // cap
	EnergyRegen float64 // regen per tick

	EnergySpentEpisode float64 // sum of energy actually spent during the current user line
	// last self-cleanup (prune) event (for demo/board)
	LastCleanupTick  int
	LastCleanupCount int

	// =======================
	// NEW: inhibition + prediction + error-driven learning + forgetting
	// =======================

	// Field inhibition levels by target (we mainly inhibit STRUCT names)
	Inhib      map[string]float64
	InhibDecay float64 // per tick multiplicative decay (e.g., 0.18 => ~18% drop)

	// Tick-local struct masses (for competition)
	ThisStructMass map[string]float64

	// Prediction model: for each struct -> counts of next token
	TransCounts map[string]map[string]float64
	BestPred    map[string]string
	PredConf    map[string]float64 // struct -> confidence 0..1

	// Expectations from previous tick: struct -> predicted next token
	PendingExpect map[string]string
	ThisExpect    map[string]string

	// Error-driven learning boost
	ErrTTL  int     // ticks remaining of "error context"
	ErrGain float64 // extra multiplier added to learning rates while ErrTTL>0

	// Error spam control (per-struct cooldown)
	ErrCooldown      map[string]int // struct -> ticks remaining to suppress repeated ERR emission
	ErrCooldownTicks int            // default cooldown applied after an ERR

	// Forgetting / pruning
	BlockLastFire map[string]int // block ID -> last tick it produced STRUCT or ACTION
	ForgetAfter   int            // ticks without firing => prune
	PruneEvery    int            // prune cadence
}

func NewContext() *Context {
	return &Context{
		Tick:         0,
		RecentActs:   make([]Signal, 0, 256),
		RecentStruct: make([]Signal, 0, 256),

		Blocks: make(map[string]Block),
		Order:  make([]string, 0, 256),

		Sensors: make(map[string]bool),

		SeenPairs:    make(map[string]float64),
		SeenComposes: make(map[string]float64),
		SeenSeq:      make(map[string]float64),

		PrevSens:      "",
		LastSens:      "",
		PrevStructSet: make(map[string]bool),
		ThisStructSet: make(map[string]bool),

		LearningEnabled: true,

		Energy:             10.0,
		EnergyMax:          10.0,
		EnergyRegen:        0.8,
		EnergySpentEpisode: 0.0,

		Inhib:          make(map[string]float64),
		InhibDecay:     0.18,
		ThisStructMass: make(map[string]float64),

		TransCounts:   make(map[string]map[string]float64),
		BestPred:      make(map[string]string),
		PredConf:      make(map[string]float64),
		PendingExpect: make(map[string]string),
		ThisExpect:    make(map[string]string),
		ErrTTL:        0,
		ErrGain:       1.2,

		ErrCooldown:      make(map[string]int),
		ErrCooldownTicks: 2,

		BlockLastFire:    make(map[string]int),
		ForgetAfter:      40, // demo-friendly: blocks fade if unused for a while
		PruneEvery:       10,
		LastCleanupTick:  0,
		LastCleanupCount: 0,
	}
}

func (c *Context) AddBlock(b Block) {
	id := b.ID()
	if _, exists := c.Blocks[id]; exists {
		return
	}
	c.Blocks[id] = b
	c.Order = append(c.Order, id)

	// start life as "recently seen" (prevents immediate pruning)
	// sensors never fire STRUCT/ACTION, but we keep them anyway
	c.BlockLastFire[id] = c.Tick
}

func (c *Context) WindowTrim(maxAge int) {
	// keep only signals within last maxAge ticks
	cutAct := 0
	for cutAct < len(c.RecentActs) && c.RecentActs[cutAct].Time < c.Tick-maxAge {
		cutAct++
	}
	if cutAct > 0 {
		c.RecentActs = append(c.RecentActs[:0], c.RecentActs[cutAct:]...)
	}

	cutSt := 0
	for cutSt < len(c.RecentStruct) && c.RecentStruct[cutSt].Time < c.Tick-maxAge {
		cutSt++
	}
	if cutSt > 0 {
		c.RecentStruct = append(c.RecentStruct[:0], c.RecentStruct[cutSt:]...)
	}
}

func minStr(a, b string) string {
	if a < b {
		return a
	}
	return b
}
func maxStr(a, b string) string {
	if a > b {
		return a
	}
	return b
}
func canonicalPairName(a, b string) string {
	lo := minStr(a, b)
	hi := maxStr(a, b)
	return fmt.Sprintf("[%s-%s]", lo, hi)
}
func pairKey(a, b string) string {
	// stable key (unordered)
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func parsePairMembers(base string) (a, b string, ok bool) {
	// expects base like "[1-2]" (exactly one dash, starts '[' ends ']')
	if len(base) < 5 || base[0] != '[' || base[len(base)-1] != ']' {
		return "", "", false
	}
	inner := base[1 : len(base)-1]
	parts := strings.Split(inner, "-")
	if len(parts) != 2 {
		return "", "", false
	}
	a = strings.TrimSpace(parts[0])
	b = strings.TrimSpace(parts[1])
	if a == "" || b == "" {
		return "", "", false
	}
	return a, b, true
}

// ---- Sensor block ----

type SensorBlock struct {
	token string
}

func (b *SensorBlock) ID() string { return "SENSOR:" + b.token }

func (b *SensorBlock) React(s Signal, ctx *Context) []Signal {
	if s.Kind == K_SENS && s.Value == b.token {
		return []Signal{{
			Kind:  K_ACT,
			Value: b.token,
			Mass:  s.Mass,
			Time:  ctx.Tick,
			From:  b.ID(),
		}}
	}
	return nil
}
func (b *SensorBlock) Tick(ctx *Context) []Signal { return nil }

// ---- Pair block: emits STRUCT([a-b]) when both ACTs are near in time ----

type CoActBlock struct {
	a, b         string
	name         string
	accum        float64
	threshold    float64
	window       int
	decayPerTick float64
	emitMass     float64
}

func NewCoActBlock(a, b string) *CoActBlock {
	lo := minStr(a, b)
	hi := maxStr(a, b)
	name := fmt.Sprintf("[%s-%s]", lo, hi)

	return &CoActBlock{
		a:            lo,
		b:            hi,
		name:         name,
		accum:        0,
		threshold:    2.0,
		window:       2,
		decayPerTick: 0.15,
		emitMass:     1.0,
	}
}

func (b *CoActBlock) ID() string { return "COACT:" + b.name }

func (b *CoActBlock) React(s Signal, ctx *Context) []Signal {
	if s.Kind != K_ACT {
		return nil
	}
	if s.Value != b.a && s.Value != b.b {
		return nil
	}

	other := b.b
	if s.Value == b.b {
		other = b.a
	}

	for i := len(ctx.RecentActs) - 1; i >= 0; i-- {
		r := ctx.RecentActs[i]
		if r.Time < ctx.Tick-b.window {
			break
		}
		if r.Kind == K_ACT && r.Value == other {
			b.accum += 1.0
			break
		}
	}

	if b.accum >= b.threshold {
		b.accum = b.threshold * 0.5
		return []Signal{{
			Kind:  K_STRUCT,
			Value: b.name,
			Mass:  b.emitMass,
			Time:  ctx.Tick,
			From:  b.ID(),
		}}
	}
	return nil
}

func (b *CoActBlock) Tick(ctx *Context) []Signal {
	if b.accum > 0 {
		b.accum -= b.decayPerTick
		if b.accum < 0 {
			b.accum = 0
		}
	}
	return nil
}

// ---- Seq block: emits STRUCT((a>b)) when a then b occurred (directional) ----

type SeqBlock struct {
	a, b         string
	name         string
	accum        float64
	threshold    float64
	window       int
	decayPerTick float64
	emitMass     float64
}

func NewSeqBlock(a, b string) *SeqBlock {
	name := fmt.Sprintf("(%s>%s)", a, b)
	return &SeqBlock{
		a:            a,
		b:            b,
		name:         name,
		accum:        0,
		threshold:    2.0,
		window:       2,
		decayPerTick: 0.18,
		emitMass:     1.0,
	}
}

func (b *SeqBlock) ID() string { return "SEQ:" + b.name }

func (b *SeqBlock) React(s Signal, ctx *Context) []Signal {
	if s.Kind != K_ACT || s.Value != b.b {
		return nil
	}

	// look back for ACT(a) within window BEFORE this ACT(b)
	for i := len(ctx.RecentActs) - 1; i >= 0; i-- {
		r := ctx.RecentActs[i]
		if r.Time < ctx.Tick-b.window {
			break
		}
		if r.Kind == K_ACT && r.Value == b.a {
			b.accum += 1.0
			break
		}
	}

	if b.accum >= b.threshold {
		b.accum = b.threshold * 0.5
		return []Signal{{
			Kind:  K_STRUCT,
			Value: b.name,
			Mass:  b.emitMass,
			Time:  ctx.Tick,
			From:  b.ID(),
		}}
	}
	return nil
}

func (b *SeqBlock) Tick(ctx *Context) []Signal {
	if b.accum > 0 {
		b.accum -= b.decayPerTick
		if b.accum < 0 {
			b.accum = 0
		}
	}
	return nil
}

// ---- Compose block: emits STRUCT([[base]-x]) when base STRUCT and ACT(x) co-occur near ----

type ComposeBlock struct {
	base         string
	x            string
	name         string
	accum        float64
	threshold    float64
	window       int
	decayPerTick float64
	emitMass     float64
}

func NewComposeBlock(base, x string) *ComposeBlock {
	name := fmt.Sprintf("[%s-%s]", base, x) // base already like "[1-2]" => produces "[[1-2]-3]"

	return &ComposeBlock{
		base:         base,
		x:            x,
		name:         name,
		accum:        0,
		threshold:    4.0,
		window:       3,
		decayPerTick: 0.12,
		emitMass:     1.0,
	}
}

func (b *ComposeBlock) ID() string { return "COMPOSE:" + b.name }

func (b *ComposeBlock) React(s Signal, ctx *Context) []Signal {
	if s.Kind != K_STRUCT && s.Kind != K_ACT {
		return nil
	}

	triggered := false

	if s.Kind == K_STRUCT && s.Value == b.base {
		for i := len(ctx.RecentActs) - 1; i >= 0; i-- {
			r := ctx.RecentActs[i]
			if r.Time < ctx.Tick-b.window {
				break
			}
			if r.Kind == K_ACT && r.Value == b.x {
				triggered = true
				break
			}
		}
	} else if s.Kind == K_ACT && s.Value == b.x {
		for i := len(ctx.RecentStruct) - 1; i >= 0; i-- {
			r := ctx.RecentStruct[i]
			if r.Time < ctx.Tick-b.window {
				break
			}
			if r.Kind == K_STRUCT && r.Value == b.base {
				triggered = true
				break
			}
		}
	}

	if triggered {
		b.accum += 1.0
		if b.accum >= b.threshold {
			b.accum = b.threshold * 0.5
			return []Signal{{
				Kind:  K_STRUCT,
				Value: b.name,
				Mass:  b.emitMass,
				Time:  ctx.Tick,
				From:  b.ID(),
			}}
		}
	}

	return nil
}

func (b *ComposeBlock) Tick(ctx *Context) []Signal {
	if b.accum > 0 {
		b.accum -= b.decayPerTick
		if b.accum < 0 {
			b.accum = 0
		}
	}
	return nil
}

// ---- Action block: turns learned structures into visible actions ----

type ActionBlock struct {
	targetStruct string
	actionName   string
	accum        float64
	threshold    float64
	decayPerTick float64
}

func NewActionBlock(targetStruct, actionName string) *ActionBlock {
	return &ActionBlock{
		targetStruct: targetStruct,
		actionName:   actionName,
		accum:        0,
		threshold:    2.0,
		decayPerTick: 0.20,
	}
}

func (b *ActionBlock) ID() string { return "ACTIONBLOCK:" + b.actionName + "<-" + b.targetStruct }

func (b *ActionBlock) React(s Signal, ctx *Context) []Signal {
	if s.Kind == K_STRUCT && s.Value == b.targetStruct {
		b.accum += s.Mass
		if b.accum >= b.threshold {
			b.accum = b.threshold * 0.5
			return []Signal{{
				Kind:  K_ACTION,
				Value: b.actionName,
				Mass:  1.0,
				Time:  ctx.Tick,
				From:  b.ID(),
			}}
		}
	}
	return nil
}

func (b *ActionBlock) Tick(ctx *Context) []Signal {
	if b.accum > 0 {
		b.accum -= b.decayPerTick
		if b.accum < 0 {
			b.accum = 0
		}
	}
	return nil
}

// ---- Field helpers: inhibition + pruning + prediction bookkeeping ----

func applyInhibition(ctx *Context, s Signal) Signal {
	// ---- energy budget: STRUCT/ACTION consume field resource ----
	if s.Kind == K_STRUCT || s.Kind == K_ACTION {
		var cost float64
		if s.Kind == K_STRUCT {
			cost = 0.6
		} else { // ACTION
			cost = 0.8
		}

		if ctx.Energy < cost {
			// not enough energy: strongly dampen this signal
			s.Mass *= 0.3
		} else {
			ctx.Energy -= cost
			ctx.EnergySpentEpisode += cost
		}
	}

	// ---- error context: make STRUCT/PRED more salient (visible reflex) ----
	// IMPORTANT: do NOT boost ACTION to avoid chaotic behavior.
	if ctx.ErrTTL > 0 && (s.Kind == K_STRUCT || s.Kind == K_PRED) {
		s.Mass *= (1.0 + ctx.ErrGain*0.5)
	}

	// NOTE: inhibition intentionally applies only to ACT/STRUCT
	if s.Kind != K_ACT && s.Kind != K_STRUCT {
		return s
	}

	lvl := ctx.Inhib[s.Value]
	if lvl <= 0 {
		return s
	}

	s.Mass = s.Mass / (1.0 + lvl)
	return s
}

func decayInhibition(ctx *Context) {
	if len(ctx.Inhib) == 0 {
		return
	}
	f := 1.0 - ctx.InhibDecay
	if f < 0.0 {
		f = 0.0
	}
	for k, v := range ctx.Inhib {
		nv := v * f
		if nv < 0.02 {
			delete(ctx.Inhib, k)
		} else {
			ctx.Inhib[k] = nv
		}
	}
}

func decayErrCooldown(ctx *Context) {
	if len(ctx.ErrCooldown) == 0 {
		return
	}
	for k, v := range ctx.ErrCooldown {
		if v <= 1 {
			delete(ctx.ErrCooldown, k)
		} else {
			ctx.ErrCooldown[k] = v - 1
		}
	}
}

func argmaxMap(m map[string]float64) (bestKey string, bestVal float64, ok bool) {
	first := true
	for k, v := range m {
		if first || v > bestVal {
			bestKey, bestVal = k, v
			first = false
		}
	}
	if first {
		return "", 0, false
	}
	return bestKey, bestVal, true
}

// Forgetting: prune old learned blocks to avoid infinite growth.
// We prune: COACT, COMPOSE, ACTIONBLOCK.
// Sensors are never pruned.
func pruneOldBlocks(ctx *Context) {
	if ctx.ForgetAfter <= 0 {
		return
	}

	kill := make(map[string]bool)

	// mark old blocks
	for id, last := range ctx.BlockLastFire {
		age := ctx.Tick - last
		if age < ctx.ForgetAfter {
			continue
		}
		if strings.HasPrefix(id, "COACT:") ||
			strings.HasPrefix(id, "SEQ:") ||
			strings.HasPrefix(id, "COMPOSE:") ||
			strings.HasPrefix(id, "ACTIONBLOCK:") {
			kill[id] = true
		}
	}

	// if we remove a STRUCT-producing block, also remove action links that point to it
	// (ACTIONBLOCK IDs contain "<-<struct>")
	if len(kill) > 0 {
		for id := range ctx.Blocks {
			if !strings.HasPrefix(id, "ACTIONBLOCK:") {
				continue
			}
			// find target struct
			parts := strings.Split(id, "<-")
			if len(parts) != 2 {
				continue
			}
			target := parts[1]
			// if any COACT/COMPOSE for that target was killed, kill this actionblock too
			if kill["COACT:"+target] || kill["SEQ:"+target] || kill["COMPOSE:"+target] {
				kill[id] = true
			}

		}
	}

	if len(kill) == 0 {
		return
	}

	// delete blocks
	for id := range kill {
		delete(ctx.Blocks, id)
		delete(ctx.BlockLastFire, id)
	}
	ctx.LastCleanupTick = ctx.Tick
	ctx.LastCleanupCount = len(kill)

	// rebuild order
	newOrder := make([]string, 0, len(ctx.Order))
	for _, id := range ctx.Order {
		if _, dead := kill[id]; dead {
			continue
		}
		newOrder = append(newOrder, id)
	}
	ctx.Order = newOrder
}

// ---- Field runner (global broadcast) ----

func RunTick(ctx *Context, incoming []Signal) []Signal {
	ctx.Tick++
	ctx.WindowTrim(12)

	// field energy regenerates each tick
	ctx.Energy += ctx.EnergyRegen
	if ctx.Energy > ctx.EnergyMax {
		ctx.Energy = ctx.EnergyMax
	}

	// decay inhibition + error TTL + error cooldown
	decayInhibition(ctx)
	decayErrCooldown(ctx)
	if ctx.ErrTTL > 0 {
		ctx.ErrTTL--
	}

	// Reset per-tick collectors (no reallocation)
	clearBoolMap(ctx.ThisStructSet)
	clearFloatMap(ctx.ThisStructMass)
	clearStringMap(ctx.ThisExpect)

	// --- Prediction error check: compare incoming sensory token against expectations from previous tick
	errSignals := make([]Signal, 0, 4)
	for _, s := range incoming {
		if s.Kind != K_SENS {
			continue
		}
		actual := s.Value

		for st, pred := range ctx.PendingExpect {
			if pred == "" {
				continue
			}
			if pred != actual {
				// avoid spamming repeated errors for the same struct
				if ctx.ErrCooldown[st] > 0 {
					continue
				}

				// If the model already predicts the actual token, don't emit an error.
				if best := ctx.BestPred[st]; best != "" && best == actual {
					// also gently relieve inhibition when we're correct
					if v := ctx.Inhib[st]; v > 0 {
						ctx.Inhib[st] = v * 0.5
						if ctx.Inhib[st] < 0.02 {
							delete(ctx.Inhib, st)
						}
					}
					continue
				}

				// start cooldown on first visible error
				ctx.ErrCooldown[st] = ctx.ErrCooldownTicks

				// emit error signal
				ev := fmt.Sprintf("%s:%s->%s", st, pred, actual)
				errSignals = append(errSignals, Signal{
					Kind:  K_ERR,
					Value: ev,
					Mass:  1.0,
					Time:  ctx.Tick,
					From:  "FIELD:PRED",
				})

				// boost learning for a short window (only impacts TRAIN plasticity)
				ctx.ErrTTL = 6

				// inhibit the mispredicting struct a bit (competition/suppression)
				ctx.Inhib[st] += 0.9
			}
		}
	}

	// Update sensory adjacency memory from incoming sensory pulses
	for _, s := range incoming {
		if s.Kind == K_SENS {
			ctx.PrevSens = ctx.LastSens
			ctx.LastSens = s.Value
		}
	}

	// Step 1: blocks time dynamics
	emitted := make([]Signal, 0, 128)
	for _, id := range ctx.Order {
		out := ctx.Blocks[id].Tick(ctx)
		if len(out) > 0 {
			emitted = append(emitted, out...)
		}
	}

	// Step 2: broadcast incoming + emitted + prediction errors (bounded)
	queue := append([]Signal{}, incoming...)
	queue = append(queue, errSignals...)
	queue = append(queue, emitted...)

	// ---- emit weak PRED signals into the field (always-on expectations) ----
	// IMPORTANT: field dynamics live in the local queue. So we append weak predictions into queue.
	for st, tok := range ctx.BestPred {
		conf := ctx.PredConf[st]
		if tok == "" || conf < 0.25 {
			continue // too weak to matter
		}

		// weak predictive signal: keep same "STRUCT->TOKEN" formatting as main PRED
		ps := Signal{
			Kind:  K_PRED,
			Value: fmt.Sprintf("%s->%s", st, tok),
			Mass:  0.25 * conf, // deliberately weak
			Time:  ctx.Tick,
			From:  "FIELD:MODEL_WEAK",
		}

		// IMPORTANT: do NOT call applyInhibition here.
		// It will be applied once inside the rounds loop (single consistent field rule).
		if ps.Mass > 0.05 {
			queue = append(queue, ps)
		}
	}

	const rounds = 4
	allOut := make([]Signal, 0, 256)

	for r := 0; r < rounds; r++ {
		if len(queue) == 0 {
			break
		}

		nextQueue := make([]Signal, 0, 256)
		for _, raw := range queue {
			// field inhibition applies to ACT/STRUCT (and energy budget inside applyInhibition)
			s := applyInhibition(ctx, raw)

			// record windows for downstream blocks
			if s.Kind == K_ACT {
				ctx.RecentActs = append(ctx.RecentActs, s)
			}

			if s.Kind == K_STRUCT {
				ctx.RecentStruct = append(ctx.RecentStruct, s)
				ctx.ThisStructSet[s.Value] = true
				ctx.ThisStructMass[s.Value] += s.Mass

				// if structure is strongly inhibited, freeze its prediction temporarily
				if ctx.Inhib[s.Value] > 0.7 {
					// do not emit/store prediction while it's suppressed
					continue
				}

				// if we have a prediction for this struct, emit it as a signal and store expectation for next tick
				if pred := ctx.BestPred[s.Value]; pred != "" {
					ctx.ThisExpect[s.Value] = pred
					nextQueue = append(nextQueue, Signal{
						Kind:  K_PRED,
						Value: fmt.Sprintf("%s->%s", s.Value, pred),
						Mass:  0.6,
						Time:  ctx.Tick,
						From:  "FIELD:MODEL",
					})
				}

				// mark the producing block as alive (forgetting uses this)
				ctx.BlockLastFire[s.From] = ctx.Tick
			}

			if s.Kind == K_ACTION {
				ctx.BlockLastFire[s.From] = ctx.Tick
			}

			// broadcast to all blocks
			for _, id := range ctx.Order {
				out := ctx.Blocks[id].React(s, ctx)
				if len(out) > 0 {
					nextQueue = append(nextQueue, out...)
				}
			}

			allOut = append(allOut, s)
		}

		queue = nextQueue
	}

	// Step 2.5: Competition (inhibition) among simultaneously active STRUCTs.
	// Winner stays; losers get inhibited for upcoming ticks.
	if len(ctx.ThisStructMass) > 1 {
		winner, wMass, ok := argmaxMap(ctx.ThisStructMass)
		if ok && winner != "" {
			for st, mass := range ctx.ThisStructMass {
				if st == winner {
					continue
				}

				// inhibit weaker structures more strongly when gap is big
				delta := wMass - mass
				add := 0.7
				if delta > 0.5 {
					add = 1.0
				}
				ctx.Inhib[st] += add

				allOut = append(allOut, Signal{
					Kind:  K_INHIB,
					Value: st,
					Mass:  ctx.Inhib[st],
					Time:  ctx.Tick,
					From:  "FIELD:COMPETE",
				})
			}
		}
	}

	// Step 3: plasticity ONLY in TRAIN mode
	if ctx.LearningEnabled {
		Plasticity(ctx)
	}

	// Prediction expectations for next tick (copy, don't alias maps)
	clearStringMap(ctx.PendingExpect)
	for k, v := range ctx.ThisExpect {
		ctx.PendingExpect[k] = v
	}

	// Shift structure memory for next tick (copy, don't alias maps)
	clearBoolMap(ctx.PrevStructSet)
	for k := range ctx.ThisStructSet {
		ctx.PrevStructSet[k] = true
	}

	// Step 4: forgetting / pruning (in both modes)
	if ctx.PruneEvery > 0 && (ctx.Tick%ctx.PruneEvery == 0) {
		pruneOldBlocks(ctx)
	}
	if ctx.LastCleanupTick == ctx.Tick && ctx.LastCleanupCount > 0 {
		fmt.Printf("t=%03d FORGET: pruned %d blocks (idle>%d)\n", ctx.Tick, ctx.LastCleanupCount, ctx.ForgetAfter)
	}

	return allOut
}

// ---- Plasticity (clean): learn only adjacent pairs + struct(prevTick)->token
// + NEW: error-driven learning boost (ErrTTL)
// + NEW: prediction-model learning (struct -> next token transition counts)
// ----

func Plasticity(ctx *Context) {
	boost := 1.0
	if ctx.ErrTTL > 0 {
		boost = 1.0 + ctx.ErrGain
	}

	// ----- PAIRS: learn only adjacent tokens (PrevSens -> LastSens) -----
	if ctx.PrevSens != "" && ctx.LastSens != "" && ctx.PrevSens != ctx.LastSens {
		k := pairKey(ctx.PrevSens, ctx.LastSens)
		ctx.SeenPairs[k] += 0.40 * boost

		if ctx.SeenPairs[k] >= 1.0 {
			name := canonicalPairName(ctx.PrevSens, ctx.LastSens)
			id := "COACT:" + name
			if _, exists := ctx.Blocks[id]; !exists {
				ctx.AddBlock(NewCoActBlock(ctx.PrevSens, ctx.LastSens))
				fmt.Printf("+++ LEARNED NEW PAIR BLOCK %s\n", name)

				actName := "ACT_ON_" + name
				ctx.AddBlock(NewActionBlock(name, actName))
				fmt.Printf("+++ ATTACHED ACTION %s <- %s\n", actName, name)
			}
			ctx.SeenPairs[k] = 0.6
		}
	}

	// ----- SEQ: learn directional adjacency (PrevSens -> LastSens) -----
	if ctx.PrevSens != "" && ctx.LastSens != "" && ctx.PrevSens != ctx.LastSens {
		sk := ctx.PrevSens + ">" + ctx.LastSens
		ctx.SeenSeq[sk] += 0.45 * boost

		if ctx.SeenSeq[sk] >= 1.0 {
			name := fmt.Sprintf("(%s>%s)", ctx.PrevSens, ctx.LastSens)
			id := "SEQ:" + name
			if _, exists := ctx.Blocks[id]; !exists {
				ctx.AddBlock(NewSeqBlock(ctx.PrevSens, ctx.LastSens))
				fmt.Printf("+++ LEARNED NEW SEQ BLOCK %s\n", name)

				actName := "ACT_ON_" + name
				ctx.AddBlock(NewActionBlock(name, actName))
				fmt.Printf("+++ ATTACHED ACTION %s <- %s\n", actName, name)
			}
			ctx.SeenSeq[sk] = 0.6
		}
	}

	// ----- COMPOSE: learn only when a structure was active on previous tick, then current token arrives -----
	// CLEAN RULE: don't learn "self-transitions" back into the same pair tokens (prevents [[[1-2]]-1] noise)
	if ctx.LastSens != "" && len(ctx.PrevStructSet) > 0 {
		for base := range ctx.PrevStructSet {

			// Only allow compose from SIMPLE PAIRS like [1-2]
			a, b, ok := parsePairMembers(base)
			if !ok {
				continue
			}
			// Skip if next token is one of the pair members (prevents oscillation artifacts)
			if ctx.LastSens == a || ctx.LastSens == b {
				continue
			}

			ck := base + "||" + ctx.LastSens
			ctx.SeenComposes[ck] += 0.28 * boost

			if ctx.SeenComposes[ck] >= 1.0 {
				name := fmt.Sprintf("[%s-%s]", base, ctx.LastSens)
				id := "COMPOSE:" + name
				if _, exists := ctx.Blocks[id]; !exists {
					ctx.AddBlock(NewComposeBlock(base, ctx.LastSens))
					fmt.Printf("+++ LEARNED NEW COMPOSE BLOCK %s\n", name)

					actName := "ACT_ON_" + name
					ctx.AddBlock(NewActionBlock(name, actName))
					fmt.Printf("+++ ATTACHED ACTION %s <- %s\n", actName, name)
				}
				ctx.SeenComposes[ck] = 0.6
			}
		}
	}

	// ----- PREDICTION MODEL LEARNING (struct prev tick -> current token) -----
	// Reinforce: if a struct was active on previous tick and token arrives now,
	// we reinforce that transition. Error context boosts reinforcement.
	if ctx.LastSens != "" && len(ctx.PrevStructSet) > 0 {
		for st := range ctx.PrevStructSet {
			if _, ok := ctx.TransCounts[st]; !ok {
				ctx.TransCounts[st] = make(map[string]float64)
			}
			ctx.TransCounts[st][ctx.LastSens] += 0.65 * boost

			// recompute best prediction + confidence for st
			bestTok := ""
			bestV := -1.0
			sumV := 0.0

			for tok, v := range ctx.TransCounts[st] {
				sumV += v
				if v > bestV {
					bestV = v
					bestTok = tok
				}
			}

			oldPred := ctx.BestPred[st]
			oldConf := ctx.PredConf[st]

			if bestTok != "" {
				ctx.BestPred[st] = bestTok
				if sumV > 0 {
					ctx.PredConf[st] = bestV / sumV
				} else {
					ctx.PredConf[st] = 0
				}
			}

			// log when prediction changes or confidence jumps
			if ctx.BestPred[st] != "" &&
				(ctx.BestPred[st] != oldPred || (ctx.PredConf[st]-oldConf) > 0.15) {

				if showPredEvents {
					fmt.Printf(
						"+++ EMERGENT EXPECTATION: %s ⇒ %s (stability=%.2f)\n",
						st,
						ctx.BestPred[st],
						ctx.PredConf[st],
					)

				}
			}
		}
	}
}

// ---- Investor-friendly BOARD ----

func resetEpisodeBoundary(ctx *Context) {
	// Prevent learning pairs across separate user lines
	ctx.PrevSens, ctx.LastSens = "", ""

	// Clear tick-local struct sets so compose doesn't leak across lines
	clearBoolMap(ctx.PrevStructSet)
	clearBoolMap(ctx.ThisStructSet)

	// Clear pending expectations to avoid “cross-episode” prediction errors
	clearStringMap(ctx.PendingExpect)

	// NEW: clear short-term field windows to prevent cross-episode "ghost" activations
	ctx.RecentActs = ctx.RecentActs[:0]
	ctx.RecentStruct = ctx.RecentStruct[:0]
	ctx.EnergySpentEpisode = 0
}

func uniqueSorted(xs []string) []string {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	out := make([]string, 0, len(m))
	for x := range m {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

func blockNamesByPrefixLast(ctx *Context, prefix string, n int) []string {
	found := make([]string, 0, n)
	for i := len(ctx.Order) - 1; i >= 0 && len(found) < n; i-- {
		id := ctx.Order[i]
		if strings.HasPrefix(id, prefix) {
			found = append(found, strings.TrimPrefix(id, prefix))
		}
	}
	for i, j := 0, len(found)-1; i < j; i, j = i+1, j-1 {
		found[i], found[j] = found[j], found[i]
	}
	return found
}

func countBlocksByPrefix(ctx *Context, prefix string) int {
	c := 0
	for id := range ctx.Blocks {
		if strings.HasPrefix(id, prefix) {
			c++
		}
	}
	return c
}

func topInhibitions(ctx *Context, n int) []string {
	type kv struct {
		k string
		v float64
	}
	arr := make([]kv, 0, len(ctx.Inhib))
	for k, v := range ctx.Inhib {
		arr = append(arr, kv{k: k, v: v})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].v > arr[j].v })
	if len(arr) > n {
		arr = arr[:n]
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		out = append(out, fmt.Sprintf("%s:%.2f", it.k, it.v))
	}
	return out
}

func topPredictions(ctx *Context, structs []string, n int) []string {
	uniq := uniqueSorted(structs)
	out := make([]string, 0, n)
	for _, st := range uniq {
		if len(out) >= n {
			break
		}
		if pred := ctx.BestPred[st]; pred != "" {
			conf := ctx.PredConf[st]
			out = append(out, fmt.Sprintf("%s⇒%s(st=%.2f)", st, pred, conf))
		}

	}
	return out
}

func printBoard(ctx *Context, episodeStructs, episodeActions, episodeErrs []string) {
	mode := "TRAIN"
	if !ctx.LearningEnabled {
		mode = "TEST"
	}

	pairs := countBlocksByPrefix(ctx, "COACT:")
	comps := countBlocksByPrefix(ctx, "COMPOSE:")
	acts := countBlocksByPrefix(ctx, "ACTIONBLOCK:")

	fmt.Printf("=== BOARD t=%03d mode=%s ===\n", ctx.Tick, mode)
	fmt.Printf("LEARNED: pairs=%d composes=%d actionLinks=%d blocks=%d\n", pairs, comps, acts, len(ctx.Blocks))
	fmt.Printf("FIELD: energy=%.2f/%.2f\n", ctx.Energy, ctx.EnergyMax)
	fmt.Printf("FIELD: energy_spent_episode=%.2f\n", ctx.EnergySpentEpisode)
	if ctx.LastCleanupTick == ctx.Tick && ctx.LastCleanupCount > 0 {
		fmt.Printf("SELF-CLEANUP: pruned %d inactive hypotheses\n", ctx.LastCleanupCount)
	}

	lastPairs := blockNamesByPrefixLast(ctx, "COACT:", 5)
	lastComps := blockNamesByPrefixLast(ctx, "COMPOSE:", 5)

	if len(lastPairs) > 0 {
		fmt.Printf("LAST PAIRS:   %v\n", lastPairs)
	}
	if len(lastComps) > 0 {
		fmt.Printf("LAST COMPOSE: %v\n", lastComps)
	}

	es := uniqueSorted(episodeStructs)
	ea := uniqueSorted(episodeActions)
	ee := uniqueSorted(episodeErrs)

	if len(es) == 0 {
		fmt.Println("EPISODE: structs=(none)")
	} else {
		fmt.Printf("EPISODE: structs=%v\n", es)
	}
	if len(ea) == 0 {
		fmt.Println("EPISODE: actions=(none)")
	} else {
		fmt.Printf("EPISODE: actions=%v\n", ea)
	}
	if len(ee) == 0 {
		fmt.Println("EPISODE: errors=(none)")
	} else {
		derrs := make([]string, 0, len(ee))
		for _, ev := range ee {
			derrs = append(derrs, strings.ReplaceAll(ev, "->", "⇒"))
		}
		fmt.Printf("EPISODE: errors=%v\n", derrs)
	}

	// show predictions for active structs
	preds := topPredictions(ctx, episodeStructs, 6)
	if len(preds) == 0 {
		fmt.Println("FIELD: dominant expectations=(none)")
	} else {
		fmt.Printf("FIELD: dominant expectations=%v\n", preds)
	}

	// ---- ALWAYS-ON expectations (all), even if no STRUCT fired this episode ----
	all := make([]string, 0, 8)
	for st, tok := range ctx.BestPred {
		if tok == "" {
			continue
		}
		conf := ctx.PredConf[st]
		if conf < 0.25 {
			continue
		}
		all = append(all, fmt.Sprintf("%s⇒%s(st=%.2f)", st, tok, conf))
	}
	sort.Strings(all)

	if len(all) == 0 {
		fmt.Println("FIELD: all expectations=(none)")
	} else {
		// ограничим вывод, чтобы не спамить
		if len(all) > 8 {
			all = all[:8]
		}

		// If ALL == DOMINANT, hide ALL to avoid looking like a duplicated/broken metric.
		if strings.Join(all, "|") != strings.Join(preds, "|") {
			fmt.Printf("FIELD: all expectations=%v\n", all)
		}
	}

	// show inhibition snapshot
	inhs := topInhibitions(ctx, 6)
	if len(inhs) == 0 {
		fmt.Println("FIELD: inhib=(none)")
	} else {
		fmt.Printf("FIELD: inhib=%v\n", inhs)
	}

	// show whether error-driven boost is active
	if ctx.ErrTTL > 0 {
		fmt.Printf("LEARNING: error-boost=ON ttl=%d gain=%.2f\n", ctx.ErrTTL, ctx.ErrGain)
	} else {
		fmt.Println("LEARNING: error-boost=OFF")
	}
}

func parseErrTriplet(ev string) (st, pred, actual string, ok bool) {
	// ev format: "<st>:<pred>-><actual>" (internal), rendered as "⇒" in logs if needed
	i := strings.Index(ev, ":")
	j := strings.LastIndex(ev, "->")
	if i <= 0 || j <= i+1 || j+2 >= len(ev) {
		return "", "", "", false
	}
	st = ev[:i]
	pred = ev[i+1 : j]
	actual = ev[j+2:]
	if st == "" || pred == "" || actual == "" {
		return "", "", "", false
	}
	return st, pred, actual, true
}

// ---- main ----

func main() {
	ctx := NewContext()
	sleepMs := 12 // demo pacing; set 0 for max speed
	autoBoard := true
	investorMode := false // concise logs, no autoBoard

	// Cosmetic-only: used to make demo look like a presentation
	demoRunning := false

	fmt.Println("STB DEMO (INHIB+PRED+ERROR+FORGET): signals -> blocks -> competition -> prediction -> error-driven learning -> forgetting.")
	fmt.Println("Commands: train | test | reset | board | quit")
	fmt.Println("Suggested demo:")
	fmt.Println("  train  ; reset ; repeat: 1 2 3   (3-5 times)")
	fmt.Println("  test   ; reset ; try:    1 2 4   (watch PRED and ERR + inhibition)")
	fmt.Println("Input tokens separated by spaces. Example: 1 2 1 2 1 2 3 1 2 3 1 2 4")

	in := bufio.NewScanner(os.Stdin)

	// Store last episode results so "board" can show them on demand
	lastEpisodeStructs := []string{}
	lastEpisodeActions := []string{}
	lastEpisodeErrs := []string{}

	for {
		fmt.Print("> ")
		if !in.Scan() {
			break
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		switch strings.ToLower(line) {
		case "quit", "exit":
			return
		case "train":
			ctx.LearningEnabled = true
			fmt.Println("MODE = TRAIN (learning enabled)")
			continue
		case "test":
			ctx.LearningEnabled = false
			fmt.Println("MODE = TEST (learning disabled)")
			continue
		case "reset":
			resetEpisodeBoundary(ctx)
			fmt.Println("Reset episode boundary")
			continue
		case "board":
			printBoard(ctx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs)
			continue
		case "autoboard on":
			autoBoard = true
			fmt.Println("Auto board = ON")
			continue
		case "autoboard off":
			autoBoard = false
			fmt.Println("Auto board = OFF")
			continue
		case "investor on":
			investorMode = true
			autoBoard = false
			sleepMs = 0
			fmt.Println("Investor mode = ON (concise event log, autoboard off, no sleep)")
			continue
		case "investor off":
			investorMode = false
			fmt.Println("Investor mode = OFF")
			continue
		case "predlog on":
			showPredEvents = true
			fmt.Println("Pred event log = ON")
			continue
		case "predlog off":
			showPredEvents = false
			fmt.Println("Pred event log = OFF")
			continue

		// --- C) DEMO SCRIPT ---
		case "demo":
			// Presentation setup (cosmetic only)
			demoRunning = true
			investorMode = true
			autoBoard = false
			sleepMs = 0
			fmt.Println("Demo: investor mode ON, running scripted sequence...")

			// Local helper: run exactly like normal input processing (same logic, no model changes).
			runLine := func(line string) {
				// Each line is a separate episode boundary (same as normal)
				resetEpisodeBoundary(ctx)

				episodeStructs := make([]string, 0, 32)
				episodeActions := make([]string, 0, 32)
				episodeErrs := make([]string, 0, 32)

				tokens := strings.Fields(line)
				for _, tok := range tokens {
					// AUTO-SEED SENSOR IF NEEDED
					if !ctx.Sensors[tok] {
						ctx.AddBlock(&SensorBlock{token: tok})
						ctx.Sensors[tok] = true
						// B) neutral message in TEST
						if ctx.LearningEnabled {
							fmt.Printf("+++ AUTO-SENSOR CREATED [%s]\n", tok)
						} else {
							fmt.Printf("+++ NEW TOKEN REGISTERED [%s] (no learning)\n", tok)
						}
					}

					inSig := Signal{Kind: K_SENS, Value: tok, Mass: 1.0, Time: ctx.Tick, From: "USER"}

					// snapshot BEFORE RunTick (so we can show Δconfidence on error)
					oldConf := make(map[string]float64, len(ctx.PendingExpect))
					oldBest := make(map[string]string, len(ctx.PendingExpect))
					for st := range ctx.PendingExpect {
						oldConf[st] = ctx.PredConf[st]
						oldBest[st] = ctx.BestPred[st]
					}

					// --- MASS ACCUMULATION SNAPSHOT (for CHARGE prints) ---
					oldLast := ctx.LastSens // what was last token before this input

					oldPair := 0.0
					oldSeq := 0.0
					pairName := ""
					seqName := ""
					pairK := ""
					seqK := ""

					if oldLast != "" && oldLast != tok {
						pairK = pairKey(oldLast, tok)
						oldPair = ctx.SeenPairs[pairK]
						pairName = canonicalPairName(oldLast, tok) // like [1-2]

						seqK = oldLast + ">" + tok
						oldSeq = ctx.SeenSeq[seqK]
						seqName = fmt.Sprintf("(%s>%s)", oldLast, tok)
					}

					// Compose snapshots (if previous tick had structs)
					oldCompose := make(map[string]float64, 4) // key -> old mass
					composeName := make(map[string]string, 4) // key -> pretty name
					if len(ctx.PrevStructSet) > 0 {
						for base := range ctx.PrevStructSet {
							a, b, ok := parsePairMembers(base)
							if !ok {
								continue
							}
							if tok == a || tok == b {
								continue // same rule as Plasticity
							}

							ck := base + "||" + tok
							oldCompose[ck] = ctx.SeenComposes[ck]
							composeName[ck] = fmt.Sprintf("[%s-%s]", base, tok) // e.g. [[1-2]-3]
						}
					}

					// Prediction transition snapshots (struct -> tok)
					oldTrans := make(map[string]float64, 4) // st -> old weight for this tok
					if len(ctx.PrevStructSet) > 0 {
						for st := range ctx.PrevStructSet {
							if m, ok := ctx.TransCounts[st]; ok {
								oldTrans[st] = m[tok]
							} else {
								oldTrans[st] = 0
							}
						}
					}

					out := RunTick(ctx, []Signal{inSig})

					actions := make([]string, 0, 8)
					structs := make([]string, 0, 8)
					errs := make([]string, 0, 8)

					for _, s := range out {
						if s.Kind == K_ACTION {
							actions = append(actions, s.Value)
							episodeActions = append(episodeActions, s.Value)
						}
						if s.Kind == K_STRUCT {
							structs = append(structs, s.Value)
							episodeStructs = append(episodeStructs, s.Value)
						}
						if s.Kind == K_ERR {
							errs = append(errs, s.Value)
							episodeErrs = append(episodeErrs, s.Value)
						}
					}

					// --- Build CHARGE lines (accumulation visibility) ---
					chargeLines := make([]string, 0, 8)

					if pairK != "" {
						now := ctx.SeenPairs[pairK]
						if now > oldPair {
							chargeLines = append(chargeLines,
								fmt.Sprintf("CHARGE PAIR %s  mass=%.2f/1.00 (+%.2f)", pairName, now, now-oldPair),
							)
							if now >= 0.85 && now < 1.0 {
								chargeLines = append(chargeLines,
									fmt.Sprintf("NEAR-CRYSTAL %s (next repeat likely forms a block)", pairName),
								)
							}
						}
					}
					if seqK != "" {
						now := ctx.SeenSeq[seqK]
						if now > oldSeq {
							chargeLines = append(chargeLines,
								fmt.Sprintf("CHARGE SEQ  %s  mass=%.2f/1.00 (+%.2f)", seqName, now, now-oldSeq),
							)
							if now >= 0.85 && now < 1.0 {
								chargeLines = append(chargeLines,
									fmt.Sprintf("NEAR-CRYSTAL %s", seqName),
								)
							}
						}
					}
					for ck, oldV := range oldCompose {
						now := ctx.SeenComposes[ck]
						if now > oldV {
							nm := composeName[ck]
							chargeLines = append(chargeLines,
								fmt.Sprintf("CHARGE COMP %s  mass=%.2f/1.00 (+%.2f)", nm, now, now-oldV),
							)
							if now >= 0.85 && now < 1.0 {
								chargeLines = append(chargeLines,
									fmt.Sprintf("NEAR-CRYSTAL %s", nm),
								)
							}
						}
					}
					for st, oldW := range oldTrans {
						nowW := 0.0
						if m, ok := ctx.TransCounts[st]; ok {
							nowW = m[tok]
						}
						if nowW > oldW {
							chargeLines = append(chargeLines,
								fmt.Sprintf("CHARGE PRED %s -> %s  w=%.2f (+%.2f)", st, tok, nowW, nowW-oldW),
							)
						}
					}

					// --- INVESTOR MODE printing rules ---
					if investorMode {
						if len(structs) == 0 && len(actions) == 0 && len(errs) == 0 && len(chargeLines) == 0 {
							// silent tick
							continue
						}

						if len(structs) == 0 && len(actions) == 0 && len(errs) == 0 && len(chargeLines) > 0 {
							fmt.Printf("t=%03d INPUT=%s\n", ctx.Tick, tok)
							for _, ln := range chargeLines {
								fmt.Printf("           %s\n", ln)
							}
							fmt.Printf("           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)

							lastEpisodeStructs = episodeStructs
							lastEpisodeActions = episodeActions
							lastEpisodeErrs = episodeErrs
							continue
						}
					}

					// normal print (or investor event tick)
					if len(structs) > 0 {
						fmt.Printf("t=%03d INPUT=%s  STRUCT=%v\n", ctx.Tick, tok, structs)
					} else {
						fmt.Printf("t=%03d INPUT=%s\n", ctx.Tick, tok)
					}

					if !investorMode && len(chargeLines) > 0 {
						for _, ln := range chargeLines {
							fmt.Printf("           %s\n", ln)
						}
					}

					// Cosmetic: during demo, hide action lines (actions still exist; only output is hidden)
					if len(actions) > 0 && !(investorMode && demoRunning) {
						fmt.Printf("           ACTION=%v\n", actions)
					}
					if len(errs) > 0 {
						// render collapse with STB arrow (keep internal "->" for parsing/logic)
						derrs := make([]string, 0, len(errs))
						for _, ev := range errs {
							derrs = append(derrs, strings.ReplaceAll(ev, "->", "⇒"))
						}
						fmt.Printf("           COLLAPSE=%v\n", derrs)
					}

					// collect inhibited hypotheses for a summary line (unique structs from this error burst)
					suppressed := make(map[string]bool, 8)

					// Cosmetic: one short "expected" line before detailed EXPECT prints
					if investorMode && len(errs) > 0 {
						uniq := make(map[string]bool, 8)
						parts := make([]string, 0, 8)

						for _, ev := range errs {
							st, pred, _, ok := parseErrTriplet(ev)
							if !ok || uniq[st] {
								continue
							}
							uniq[st] = true

							showPred := pred
							if b := oldBest[st]; b != "" {
								showPred = b
							}
							conf := oldConf[st]
							parts = append(parts, fmt.Sprintf("%s⇒%s(stability=%.2f)", st, showPred, conf))
						}

						if len(parts) > 0 {
							fmt.Printf("           EMERGENT EXPECTATIONS: %s\n", strings.Join(parts, " ; "))
						}
					}

					for _, ev := range errs {
						st, pred, actual, ok := parseErrTriplet(ev)
						if !ok {
							continue
						}

						if ctx.Inhib[st] > 0.0 {
							suppressed[st] = true
						}

						// confidence delta (old -> new)
						before := oldConf[st]
						after := ctx.PredConf[st]

						// if oldBest is empty, fall back to pred from the event
						showPred := pred
						if b := oldBest[st]; b != "" {
							showPred = b
						}

						if ctx.LearningEnabled {
							fmt.Printf("           EXPECTATION COLLAPSE: %s expected %s (stability %.2f) ⇒ got %s -> UPDATED (stability %.2f->%.2f)\n",
								st, showPred, before, actual, before, after)
						} else {
							fmt.Printf("           EXPECTATION COLLAPSE: %s expected %s (stability %.2f) ⇒ got %s\n",
								st, showPred, before, actual)
						}

					}

					// A) investor summary line after error burst
					if investorMode && len(errs) > 0 {
						inhibN := len(suppressed)
						if inhibN == 0 {
							inhibN = len(errs) // fallback
						}
						fmt.Printf("           FIELD RESPONSE: inhibition=%d hypotheses | error-context ttl=%d gain=%.2f\n",
							inhibN, ctx.ErrTTL, ctx.ErrGain)

					}

					// --- show energy "at the moment" of activity ---
					if len(structs) > 0 || len(actions) > 0 || len(errs) > 0 {
						fmt.Printf("           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)
					}

					// Save last episode for "board" command
					lastEpisodeStructs = episodeStructs
					lastEpisodeActions = episodeActions
					lastEpisodeErrs = episodeErrs
				}

				// Save last episode for "board" command (end of line)
				lastEpisodeStructs = episodeStructs
				lastEpisodeActions = episodeActions
				lastEpisodeErrs = episodeErrs
			}

			// --- demo sequence ---
			ctx.LearningEnabled = true
			fmt.Println("PHASE 1/3: ACCUMULATION -> CRYSTALLIZATION")
			runLine("1 2 3 1 2 3 1 2 3 1 2 3")

			fmt.Println("PHASE 2/3: STRUCTURES -> PREDICTIONS")
			runLine("1 2 3 1 2 3 1 2 3 1 2 3")
			printBoard(ctx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs)

			ctx.LearningEnabled = false
			fmt.Println("PHASE 3/3: EXPECTATION COLLAPSE -> INHIBITION (ERROR-CONTEXT)")
			runLine("1 2 4")
			runLine("1 2 4")
			printBoard(ctx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs)

			// Cosmetic: final one-shot summary (no logic change; counts from existing blocks)
			pairs := countBlocksByPrefix(ctx, "COACT:")
			comps := countBlocksByPrefix(ctx, "COMPOSE:")
			acts := countBlocksByPrefix(ctx, "ACTIONBLOCK:")

			fmt.Printf("DEMO SUMMARY: learned pairs=%d | composes=%d | actionLinks=%d | blocks=%d\n",
				pairs, comps, acts, len(ctx.Blocks))

			// keep investorMode ON after demo (intentional), but stop hiding action lines outside demo
			demoRunning = false
			continue
		}

		// Each user line is a separate episode boundary (prevents cross-line adjacency learning + prediction leakage)
		resetEpisodeBoundary(ctx)

		episodeStructs := make([]string, 0, 32)
		episodeActions := make([]string, 0, 32)
		episodeErrs := make([]string, 0, 32)

		tokens := strings.Fields(line)
		for _, tok := range tokens {
			// AUTO-SEED SENSOR IF NEEDED
			if !ctx.Sensors[tok] {
				ctx.AddBlock(&SensorBlock{token: tok})
				ctx.Sensors[tok] = true

				// B) neutral message in TEST
				if ctx.LearningEnabled {
					fmt.Printf("+++ AUTO-SENSOR CREATED [%s]\n", tok)
				} else {
					fmt.Printf("+++ NEW TOKEN REGISTERED [%s] (no learning)\n", tok)
				}
			}

			inSig := Signal{Kind: K_SENS, Value: tok, Mass: 1.0, Time: ctx.Tick, From: "USER"}

			// snapshot BEFORE RunTick (so we can show Δconfidence on error)
			oldConf := make(map[string]float64, len(ctx.PendingExpect))
			oldBest := make(map[string]string, len(ctx.PendingExpect))
			for st := range ctx.PendingExpect {
				oldConf[st] = ctx.PredConf[st]
				oldBest[st] = ctx.BestPred[st]
			}

			// --- MASS ACCUMULATION SNAPSHOT (for CHARGE prints) ---
			oldLast := ctx.LastSens // what was last token before this input

			oldPair := 0.0
			oldSeq := 0.0
			pairName := ""
			seqName := ""
			pairK := ""
			seqK := ""

			if oldLast != "" && oldLast != tok {
				pairK = pairKey(oldLast, tok)
				oldPair = ctx.SeenPairs[pairK]
				pairName = canonicalPairName(oldLast, tok) // like [1-2]

				seqK = oldLast + ">" + tok
				oldSeq = ctx.SeenSeq[seqK]
				seqName = fmt.Sprintf("(%s>%s)", oldLast, tok)
			}

			// Compose snapshots (if previous tick had structs)
			oldCompose := make(map[string]float64, 4) // key -> old mass
			composeName := make(map[string]string, 4) // key -> pretty name
			if len(ctx.PrevStructSet) > 0 {
				for base := range ctx.PrevStructSet {
					a, b, ok := parsePairMembers(base)
					if !ok {
						continue
					}
					if tok == a || tok == b {
						continue // same rule as Plasticity
					}

					ck := base + "||" + tok
					oldCompose[ck] = ctx.SeenComposes[ck]
					composeName[ck] = fmt.Sprintf("[%s-%s]", base, tok) // e.g. [[1-2]-3]
				}
			}

			// Prediction transition snapshots (struct -> tok)
			oldTrans := make(map[string]float64, 4) // st -> old weight for this tok
			if len(ctx.PrevStructSet) > 0 {
				for st := range ctx.PrevStructSet {
					if m, ok := ctx.TransCounts[st]; ok {
						oldTrans[st] = m[tok]
					} else {
						oldTrans[st] = 0
					}
				}
			}

			out := RunTick(ctx, []Signal{inSig})

			actions := make([]string, 0, 8)
			structs := make([]string, 0, 8)
			errs := make([]string, 0, 8)

			for _, s := range out {
				if s.Kind == K_ACTION {
					actions = append(actions, s.Value)
					episodeActions = append(episodeActions, s.Value)
				}
				if s.Kind == K_STRUCT {
					structs = append(structs, s.Value)
					episodeStructs = append(episodeStructs, s.Value)
				}
				if s.Kind == K_ERR {
					errs = append(errs, s.Value)
					episodeErrs = append(episodeErrs, s.Value)
				}
			}

			// --- Build CHARGE lines (accumulation visibility) ---
			chargeLines := make([]string, 0, 8)

			if pairK != "" {
				now := ctx.SeenPairs[pairK]
				if now > oldPair {
					chargeLines = append(chargeLines,
						fmt.Sprintf("CHARGE PAIR %s  mass=%.2f/1.00 (+%.2f)", pairName, now, now-oldPair),
					)
					if now >= 0.85 && now < 1.0 {
						chargeLines = append(chargeLines,
							fmt.Sprintf("NEAR-CRYSTAL %s (next repeat likely forms a block)", pairName),
						)
					}
				}
			}
			if seqK != "" {
				now := ctx.SeenSeq[seqK]
				if now > oldSeq {
					chargeLines = append(chargeLines,
						fmt.Sprintf("CHARGE SEQ  %s  mass=%.2f/1.00 (+%.2f)", seqName, now, now-oldSeq),
					)
					if now >= 0.85 && now < 1.0 {
						chargeLines = append(chargeLines,
							fmt.Sprintf("NEAR-CRYSTAL %s", seqName),
						)
					}
				}
			}
			for ck, oldV := range oldCompose {
				now := ctx.SeenComposes[ck]
				if now > oldV {
					nm := composeName[ck]
					chargeLines = append(chargeLines,
						fmt.Sprintf("CHARGE COMP %s  mass=%.2f/1.00 (+%.2f)", nm, now, now-oldV),
					)
					if now >= 0.85 && now < 1.0 {
						chargeLines = append(chargeLines,
							fmt.Sprintf("NEAR-CRYSTAL %s", nm),
						)
					}
				}
			}
			for st, oldW := range oldTrans {
				nowW := 0.0
				if m, ok := ctx.TransCounts[st]; ok {
					nowW = m[tok]
				}
				if nowW > oldW {
					chargeLines = append(chargeLines,
						fmt.Sprintf("CHARGE PRED %s -> %s  w=%.2f (+%.2f)", st, tok, nowW, nowW-oldW),
					)
				}
			}

			// --- INVESTOR MODE printing rules ---
			if investorMode {
				// If truly nothing changed, keep it silent.
				if len(structs) == 0 && len(actions) == 0 && len(errs) == 0 && len(chargeLines) == 0 {
					if sleepMs > 0 {
						time.Sleep(time.Duration(sleepMs) * time.Millisecond)
					}
					// Save last episode & continue
					lastEpisodeStructs = episodeStructs
					lastEpisodeActions = episodeActions
					lastEpisodeErrs = episodeErrs
					continue
				}

				// If only CHARGE happened, print compact “accumulation tick”
				if len(structs) == 0 && len(actions) == 0 && len(errs) == 0 && len(chargeLines) > 0 {
					fmt.Printf("t=%03d INPUT=%s\n", ctx.Tick, tok)
					for _, ln := range chargeLines {
						fmt.Printf("           %s\n", ln)
					}
					fmt.Printf("           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)

					if sleepMs > 0 {
						time.Sleep(time.Duration(sleepMs) * time.Millisecond)
					}

					// Save last episode for "board" command
					lastEpisodeStructs = episodeStructs
					lastEpisodeActions = episodeActions
					lastEpisodeErrs = episodeErrs

					// autoboard is already disabled in investor on, but keep behavior consistent
					if autoBoard {
						printBoard(ctx, episodeStructs, episodeActions, episodeErrs)
					}
					continue
				}
			}

			// normal print (or investor event tick)
			if len(structs) > 0 {
				fmt.Printf("t=%03d INPUT=%s  STRUCT=%v\n", ctx.Tick, tok, structs)
			} else {
				fmt.Printf("t=%03d INPUT=%s\n", ctx.Tick, tok)
			}

			// In verbose mode, also show CHARGE lines when present (optional, but informative)
			if !investorMode && len(chargeLines) > 0 {
				for _, ln := range chargeLines {
					fmt.Printf("           %s\n", ln)
				}
			}

			if len(actions) > 0 {
				fmt.Printf("           ACTION=%v\n", actions)
			}
			if len(errs) > 0 {
				fmt.Printf("           ERROR=%v\n", errs)
			}

			// collect inhibited hypotheses for a summary line (unique structs from this error burst)
			suppressed := make(map[string]bool, 8)

			// Cosmetic: one short "expected" line before detailed EXPECT prints
			if investorMode && len(errs) > 0 {
				uniq := make(map[string]bool, 8)
				parts := make([]string, 0, 8)

				for _, ev := range errs {
					st, pred, _, ok := parseErrTriplet(ev)
					if !ok || uniq[st] {
						continue
					}
					uniq[st] = true

					showPred := pred
					if b := oldBest[st]; b != "" {
						showPred = b
					}
					conf := oldConf[st]
					parts = append(parts, fmt.Sprintf("%s⇒%s(stability=%.2f)", st, showPred, conf))
				}

				if len(parts) > 0 {
					fmt.Printf("           EMERGENT EXPECTATIONS: %s\n", strings.Join(parts, " ; "))
				}
			}

			for _, ev := range errs {
				st, pred, actual, ok := parseErrTriplet(ev)
				if !ok {
					continue
				}

				if ctx.Inhib[st] > 0.0 {
					suppressed[st] = true
				}

				// confidence delta (old -> new)
				before := oldConf[st]
				after := ctx.PredConf[st]

				// if oldBest is empty, fall back to pred from the event
				showPred := pred
				if b := oldBest[st]; b != "" {
					showPred = b
				}

				if ctx.LearningEnabled {
					fmt.Printf("           EXPECT %s = %s (%.2f) -> GOT %s -> SURPRISE -> UPDATED (conf %.2f->%.2f)\n",
						st, showPred, before, actual, before, after)
				} else {
					fmt.Printf("           EXPECT %s = %s (%.2f) -> GOT %s -> SURPRISE\n",
						st, showPred, before, actual)
				}

				_ = after // keep "after" read; useful for future prints
			}

			// A) investor summary line after error burst
			if investorMode && len(errs) > 0 {
				inhibN := len(suppressed)
				if inhibN == 0 {
					inhibN = len(errs) // fallback
				}
				fmt.Printf("           FIELD RESPONSE: inhib=%d hypotheses | error-boost ttl=%d gain=%.2f\n",
					inhibN, ctx.ErrTTL, ctx.ErrGain)
			}

			// --- investor: show energy "at the moment" of activity ---
			if len(structs) > 0 || len(actions) > 0 || len(errs) > 0 {
				fmt.Printf("           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)
			}

			if sleepMs > 0 {
				time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			}

			// Save last episode for "board" command
			lastEpisodeStructs = episodeStructs
			lastEpisodeActions = episodeActions
			lastEpisodeErrs = episodeErrs

			// Investor-friendly board after each episode (toggleable)
			if autoBoard {
				printBoard(ctx, episodeStructs, episodeActions, episodeErrs)
			}
		}
	}
}
