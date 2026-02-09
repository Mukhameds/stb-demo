package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

var showPredEvents = true
var colorLogs = true

const (
	C_RESET = "\033[0m"
	C_BOLD  = "\033[1m"

	C_RED    = "\033[31m"
	C_GREEN  = "\033[32m"
	C_YELLOW = "\033[33m"
	C_BLUE   = "\033[34m"
	C_MAGENTA= "\033[35m"
	C_CYAN   = "\033[36m"
	C_GRAY   = "\033[90m"
)

func cwrap(s, color string) string {
	if !colorLogs || color == "" {
		return s
	}
	return color + s + C_RESET
}


func cprintf(color string, format string, args ...any) {
	if colorLogs && color != "" {
		fmt.Print(color)
	}
	fmt.Printf(format, args...)
	if colorLogs && color != "" {
		fmt.Print(C_RESET)
	}
}

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

func clearIntMap(m map[string]int) {
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



type Block interface {
	ID() string
	React(s Signal, ctx *Context) []Signal
	Tick(ctx *Context) []Signal
}



type Context struct {
	Tick         int
	RecentActs   []Signal
	RecentStruct []Signal

	Blocks map[string]Block
	Order  []string

	PredEvents  []string 
	TrainEvents []string
	LastAdapt   []string 

	
	Sensors map[string]bool

	SeenPairs    map[string]float64
	SeenComposes map[string]float64

	SeenSeq map[string]float64 

	
	PrevSens string
	LastSens string

	PrevStructSet map[string]bool 
	ThisStructSet map[string]bool 

	
	LearningEnabled bool 
	LearnStruct     bool 
	LearnPred       bool 

	DisableSeq bool 

	SuppressPredLog bool 

	
	Energy      float64 
	EnergyMax   float64 
	EnergyRegen float64 

	EnergySpentEpisode float64 
	
	LastCleanupTick  int
	LastCleanupCount int

	

	
	Inhib      map[string]float64
	InhibDecay float64 

	
	ThisStructMass map[string]float64

	
	TransCounts map[string]map[string]float64
	BestPred    map[string]string
	PredConf    map[string]float64 

	
	PendingExpect map[string]string
	ThisExpect    map[string]string

	
	ErrTTL  int     
	ErrGain float64 

	
	ErrCooldown      map[string]int 
	ErrCooldownTicks int            

	
	BlockLastFire map[string]int 
	ForgetAfter   int            
	PruneEvery    int            

	DemoFocusPairsOnly bool

	CostedThisTick map[string]bool

	ActionsThisTick int
	MaxActionsPerTick int 
	LastArmedExpect map[string]string 
	LastArmedConf   map[string]float64 
}

func NewContext() *Context {
	ctx := &Context{
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
		LearnStruct:     true,
		LearnPred:       true,

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
		LastAdapt:     make([]string, 0, 16),
		ErrTTL:        0,
		ErrGain:       1.2,

		ErrCooldown:      make(map[string]int),
		ErrCooldownTicks: 2,

		BlockLastFire: make(map[string]int),
		ForgetAfter:   120,
		PruneEvery:    20,

		LastCleanupTick:  0,
		LastCleanupCount: 0,

		MaxActionsPerTick: 1,

		DemoFocusPairsOnly: true,
		DisableSeq: false,

		
		LastArmedExpect: make(map[string]string),
		LastArmedConf:   make(map[string]float64),
	}

	return ctx
}



func (c *Context) AddBlock(b Block) {
	id := b.ID()
	if _, exists := c.Blocks[id]; exists {
		return
	}
	c.Blocks[id] = b
	c.Order = append(c.Order, id)

	
	
	c.BlockLastFire[id] = c.Tick
}

func (c *Context) WindowTrim(maxAge int) {
	
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


type CoActBlock struct {
	a, b         string
	name         string
	accum        float64
	threshold    float64
	window       int
	decayPerTick float64
	emitMass     float64
	mature bool 
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
    
    if ctx.PrevSens == "" || ctx.LastSens == "" {
        return nil
    }
    
    a := ctx.PrevSens
    c := ctx.LastSens
    if a == c {
        return nil
    }

    
    if !((a == b.a && c == b.b) || (a == b.b && c == b.a)) {
        return nil
    }

    
    ctx.BlockLastFire[b.ID()] = ctx.Tick

    if b.mature {
        // mature: emit on every adjacency match
        return []Signal{{
            Kind:  K_STRUCT,
            Value: b.name,
            Mass:  b.emitMass,
            Time:  ctx.Tick,
            From:  b.ID(),
        }}
    }

    
    b.accum += 1.0
    if b.accum >= b.threshold {
        b.accum = b.threshold * 0.5
        b.mature = true
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



type SeqBlock struct {
	a, b         string
	name         string
	accum        float64
	threshold    float64
	window       int
	decayPerTick float64
	emitMass     float64
	mature bool

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

	
	triggered := false
	adjacent := false

	for i := len(ctx.RecentActs) - 1; i >= 0; i-- {
		r := ctx.RecentActs[i]
		if r.Time < ctx.Tick-b.window {
			break
		}
		if r.Kind == K_ACT && r.Value == b.a {
			triggered = true
			if r.Time == ctx.Tick-1 {
				adjacent = true
			}
			break
		}
	}

	if !triggered {
		return nil
	}
	
	ctx.BlockLastFire[b.ID()] = ctx.Tick


	
	if b.mature {
		if !adjacent {
			return nil
		}
		return []Signal{{
			Kind:  K_STRUCT,
			Value: b.name,
			Mass:  b.emitMass,
			Time:  ctx.Tick,
			From:  b.ID(),
		}}
	}

	
	b.accum += 1.0

	if b.accum >= b.threshold {
		b.accum = b.threshold * 0.5
		b.mature = true
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
	name := fmt.Sprintf("[%s-%s]", base, x) 

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

func (ctx *Context) AllowActionThisTick() bool {
	if ctx.MaxActionsPerTick <= 0 {
		ctx.MaxActionsPerTick = 1
	}
	if ctx.ActionsThisTick >= ctx.MaxActionsPerTick {
		return false
	}
	ctx.ActionsThisTick++
	return true
}


func applyInhibition(ctx *Context, s Signal) Signal {
	
	if s.Kind == K_ACTION {
		
		if !ctx.AllowActionThisTick() {
			s.Mass = 0
			return s
		}

		
		
		key := fmt.Sprintf("%d|%d|%s|%s", ctx.Tick, s.Kind, s.Value, s.From)
		if ctx.CostedThisTick != nil && !ctx.CostedThisTick[key] {
			ctx.CostedThisTick[key] = true

			const cost = 0.8
			if ctx.Energy < cost {
				// not enough energy: dampen
				s.Mass *= 0.3
			} else {
				ctx.Energy -= cost
				ctx.EnergySpentEpisode += cost
			}
		}
	}

	
	
	if ctx.ErrTTL > 0 && (s.Kind == K_STRUCT || s.Kind == K_PRED) {
		s.Mass *= (1.0 + ctx.ErrGain*0.5)
	}

	
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
		if first ||
			v > bestVal ||
			(math.Abs(v-bestVal) < 1e-9 && preferStructName(k, bestKey)) {
			bestKey, bestVal = k, v
			first = false
		}
	}
	if first {
		return "", 0, false
	}
	return bestKey, bestVal, true
}

// 🔒 YC-determinism tie-break helper
func preferStructName(a, b string) bool {
	aIsPair := strings.HasPrefix(a, "[")
	bIsPair := strings.HasPrefix(b, "[")
	if aIsPair != bIsPair {
		return aIsPair
	}

	aIsSeq := strings.HasPrefix(a, "(")
	bIsSeq := strings.HasPrefix(b, "(")
	if aIsSeq != bIsSeq {
		return !aIsSeq
	}

	return a < b
}

func pruneOldBlocks(ctx *Context) {
	if ctx.ForgetAfter <= 0 {
		return
	}

	kill := make(map[string]bool)

	
	
	getActionTarget := func(id string) string {
		parts := strings.Split(id, "<-")
		if len(parts) != 2 {
			return ""
		}
		return parts[1]
	}

	
	
	shouldProtectStruct := func(structName string) bool {
		
		if ctx.BestPred[structName] != "" && ctx.PredConf[structName] >= 0.30 {
			return true
		}
		
		if m, ok := ctx.TransCounts[structName]; ok && len(m) > 0 {
			
			for _, w := range m {
				if w >= 0.20 {
					return true
				}
			}
		}
		
		if ctx.ThisStructSet[structName] || ctx.PrevStructSet[structName] {
			return true
		}
		return false
	}

	
	
	getStructFromID := func(id string) string {
		if strings.HasPrefix(id, "COACT:") {
			return strings.TrimPrefix(id, "COACT:")
		}
		if strings.HasPrefix(id, "SEQ:") {
			return strings.TrimPrefix(id, "SEQ:")
		}
		if strings.HasPrefix(id, "COMPOSE:") {
			return strings.TrimPrefix(id, "COMPOSE:")
		}
		return ""
	}

	
	for id, last := range ctx.BlockLastFire {
		age := ctx.Tick - last
		if age < ctx.ForgetAfter {
			continue
		}

		
		if strings.HasPrefix(id, "COACT:") ||
			strings.HasPrefix(id, "SEQ:") ||
			strings.HasPrefix(id, "COMPOSE:") {

			st := getStructFromID(id)
			if st == "" {
				continue
			}

			
			if shouldProtectStruct(st) {
				continue
			}

			kill[id] = true
		}
	}

	
	
	if len(kill) > 0 {
		for id := range ctx.Blocks {
			if !strings.HasPrefix(id, "ACTIONBLOCK:") {
				continue
			}
			target := getActionTarget(id)
			if target == "" {
				continue
			}
			// kill action only if the corresponding struct producer was killed
			if kill["COACT:"+target] || kill["SEQ:"+target] || kill["COMPOSE:"+target] {
				kill[id] = true
			}
		}
	}

	if len(kill) == 0 {
		return
	}

	
	
	const maxDeletesPerCycle = 6
	if len(kill) > maxDeletesPerCycle {
		type cand struct {
			id  string
			age int
		}
		cands := make([]cand, 0, len(kill))
		for id := range kill {
			last := ctx.BlockLastFire[id]
			cands = append(cands, cand{id: id, age: ctx.Tick - last})
		}
		sort.Slice(cands, func(i, j int) bool { return cands[i].age > cands[j].age })

		trimmed := make(map[string]bool, maxDeletesPerCycle)
		for i := 0; i < len(cands) && i < maxDeletesPerCycle; i++ {
			trimmed[cands[i].id] = true
		}
		kill = trimmed
	}

	
	for id := range kill {
		delete(ctx.Blocks, id)
		delete(ctx.BlockLastFire, id)
	}
	ctx.LastCleanupTick = ctx.Tick
	ctx.LastCleanupCount = len(kill)

	
	newOrder := make([]string, 0, len(ctx.Order))
	for _, id := range ctx.Order {
		if kill[id] {
			continue
		}
		newOrder = append(newOrder, id)
	}
	ctx.Order = newOrder
}

// -
func RunTick(ctx *Context, incoming []Signal) []Signal {
	
	if ctx.Inhib == nil {
		ctx.Inhib = make(map[string]float64)
	}
	if ctx.CostedThisTick == nil {
		ctx.CostedThisTick = make(map[string]bool)
	}
	if ctx.ErrCooldown == nil {
		ctx.ErrCooldown = make(map[string]int)
	}
	if ctx.TransCounts == nil {
		ctx.TransCounts = make(map[string]map[string]float64)
	}
	if ctx.BestPred == nil {
		ctx.BestPred = make(map[string]string)
	}
	if ctx.PredConf == nil {
		ctx.PredConf = make(map[string]float64)
	}

	
	if ctx.ThisStructSet == nil {
		ctx.ThisStructSet = make(map[string]bool)
	}
	if ctx.ThisStructMass == nil {
		ctx.ThisStructMass = make(map[string]float64)
	}
	if ctx.ThisExpect == nil {
		ctx.ThisExpect = make(map[string]string)
	}
	if ctx.PendingExpect == nil {
		ctx.PendingExpect = make(map[string]string)
	}
	if ctx.PrevStructSet == nil {
		ctx.PrevStructSet = make(map[string]bool)
	}
	if ctx.BlockLastFire == nil {
		ctx.BlockLastFire = make(map[string]int)
	}
	if ctx.LastArmedExpect == nil {
		ctx.LastArmedExpect = make(map[string]string)
	}
	if ctx.LastArmedConf == nil {
		ctx.LastArmedConf = make(map[string]float64)
	}

	
	for k := range ctx.LastArmedExpect {
		delete(ctx.LastArmedExpect, k)
	}
	for k := range ctx.LastArmedConf {
		delete(ctx.LastArmedConf, k)
	}
	for st, tok := range ctx.PendingExpect {
		if tok == "" {
			continue
		}
		ctx.LastArmedExpect[st] = tok
		ctx.LastArmedConf[st] = ctx.PredConf[st] 
	}

	
	if ctx.PredEvents != nil {
		ctx.PredEvents = ctx.PredEvents[:0]
	}

	
	ctx.ActionsThisTick = 0
	if ctx.MaxActionsPerTick <= 0 {
		ctx.MaxActionsPerTick = 1
	}

	ctx.Tick++

	
	for k := range ctx.CostedThisTick {
		delete(ctx.CostedThisTick, k)
	}

	ctx.WindowTrim(12)

	
	ctx.Energy += ctx.EnergyRegen
	if ctx.Energy > ctx.EnergyMax {
		ctx.Energy = ctx.EnergyMax
	}

	
	decayInhibition(ctx)
	decayErrCooldown(ctx)
	if ctx.ErrTTL > 0 {
		ctx.ErrTTL--
	}

	
	clearBoolMap(ctx.ThisStructSet)
	clearFloatMap(ctx.ThisStructMass)
	clearStringMap(ctx.ThisExpect)

	
	
	
	errSignals := make([]Signal, 0, 4)
	hadErrThisTick := false 
	

	
	actual := ""
	for _, s := range incoming {
		if s.Kind == K_SENS {
			actual = s.Value
			break
		}
	}

	if actual != "" {
		for st, pred := range ctx.PendingExpect {
			if pred == "" || pred == actual {
				continue
			}

			inCooldown := ctx.ErrCooldown[st] > 0

			
			errSignals = append(errSignals, Signal{
				Kind:  K_ERR,
				Value: fmt.Sprintf("%s:%s->%s", st, pred, actual),
				Mass:  1.0,
				Time:  ctx.Tick,
				From:  "FIELD:PRED",
			})

			
			if ctx.LearningEnabled && ctx.LearnPred {
				if _, ok := ctx.TransCounts[st]; !ok {
					ctx.TransCounts[st] = make(map[string]float64)
				}

			
				if w, ok := ctx.TransCounts[st][pred]; ok && w > 0 {
					ctx.TransCounts[st][pred] = w * 0.92
					if ctx.TransCounts[st][pred] < 0.05 {
						delete(ctx.TransCounts[st], pred)
					}
				}

				
				bump := 0.14
				if inCooldown {
					bump = 0.08
				}
				ctx.TransCounts[st][actual] += bump

				
				if ctx.TransCounts[st][actual] > 3.00 {
					ctx.TransCounts[st][actual] = 3.00
				}
			}

		
			if !inCooldown {
				hadErrThisTick = true

				
				ctx.ErrCooldown[st] = ctx.ErrCooldownTicks

				
				ctx.ErrTTL = 3

				
				predKey := fmt.Sprintf("%s->%s", st, pred)
				ctx.Inhib[predKey] += 0.6   // suppress wrong expectation strongly
				ctx.Inhib[st] += 0.08       // tiny damping on struct (optional, prevents runaway spam)

				
				if ctx.PredEvents != nil {
					ctx.PredEvents = append(ctx.PredEvents, fmt.Sprintf("ERR %s expected %s got %s", st, pred, actual))
				}
			}
		}
	}

	
	for _, s := range incoming {
		if s.Kind == K_SENS {
			ctx.PrevSens = ctx.LastSens
			ctx.LastSens = s.Value
		}
	}

	
	emitted := make([]Signal, 0, 128)
	for _, id := range ctx.Order {
		out := ctx.Blocks[id].Tick(ctx)
		if len(out) > 0 {
			emitted = append(emitted, out...)
		}
	}

	
	queue := append([]Signal{}, incoming...)
	queue = append(queue, errSignals...)
	queue = append(queue, emitted...)

	
	for st, tok := range ctx.BestPred {
		conf := ctx.PredConf[st]
		if tok == "" || conf < 0.25 {
			continue
		}
		ps := Signal{
			Kind:  K_PRED,
			Value: fmt.Sprintf("%s->%s", st, tok),
			Mass:  0.25 * conf,
			Time:  ctx.Tick,
			From:  "FIELD:MODEL_WEAK",
		}
		if ps.Mass > 0.05 {
			queue = append(queue, ps)
		}
	}

	const rounds = 4
	allOut := make([]Signal, 0, 256)

	for r := 0; r < rounds && len(queue) > 0; r++ {
		nextQueue := make([]Signal, 0, 256)

		for _, raw := range queue {
			s := applyInhibition(ctx, raw)
			if s.Mass <= 0 {
				continue
			}

			
			if s.Kind == K_STRUCT || s.Kind == K_ACTION || s.Kind == K_ACT {
				if s.From != "" {
					if _, ok := ctx.Blocks[s.From]; ok {
						ctx.BlockLastFire[s.From] = ctx.Tick
					}
				}
			}

			if s.Kind == K_STRUCT {
				ctx.RecentStruct = append(ctx.RecentStruct, s)

				// Core field must not UI-filter structures.
				ctx.ThisStructSet[s.Value] = true
				ctx.ThisStructMass[s.Value] += s.Mass

				if ctx.Inhib[s.Value] <= 0.7 {
					if pred := ctx.BestPred[s.Value]; pred != "" {
						nextQueue = append(nextQueue, Signal{
							Kind:  K_PRED,
							Value: fmt.Sprintf("%s->%s", s.Value, pred),
							Mass:  0.6,
							Time:  ctx.Tick,
							From:  "FIELD:MODEL",
						})
					}
				}
			}

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

	
	winner := ""
	wMass := 0.0

	if len(ctx.ThisStructMass) > 0 {
		const eps = 1e-9
		for st, mass := range ctx.ThisStructMass {
			if winner == "" {
				winner, wMass = st, mass
				continue
			}
			if mass > wMass+eps {
				winner, wMass = st, mass
				continue
			}
			if mass >= wMass-eps && preferStructName(st, winner) {
				winner, wMass = st, mass
			}
		}
	}

	if winner != "" && len(ctx.ThisStructMass) > 1 {
		for st, mass := range ctx.ThisStructMass {
			if st == winner {
				continue
			}
			add := 0.7
			if wMass-mass > 0.5 {
				add = 1.0
			}
			ctx.Inhib[st] += add
		}
	}

	
	if winner != "" {
		const structWinnerCost = 0.6
		if ctx.Energy >= structWinnerCost {
			ctx.Energy -= structWinnerCost
			ctx.EnergySpentEpisode += structWinnerCost
		} else {
			ctx.Inhib[winner] += 0.5
		}
	}

	
	if ctx.LearningEnabled {
		Plasticity(ctx, hadErrThisTick)
	}

	
	clearStringMap(ctx.ThisExpect)

	
	if !hadErrThisTick {
		if winner != "" && ctx.Inhib[winner] <= 0.7 {
			if keepPred := ctx.BestPred[winner]; keepPred != "" {
				ctx.ThisExpect[winner] = keepPred
			}
		}
	}

	
	clearStringMap(ctx.PendingExpect)
	for k, v := range ctx.ThisExpect {
		ctx.PendingExpect[k] = v
	}

	clearBoolMap(ctx.PrevStructSet)
	for k := range ctx.ThisStructSet {
		ctx.PrevStructSet[k] = true
	}

	
	if ctx.PruneEvery > 0 && ctx.Tick%ctx.PruneEvery == 0 {
		pruneOldBlocks(ctx)
	}

	

	return allOut
}


const PredSwitchMass = 1.0

func Plasticity(ctx *Context, hadErrThisTick bool) {
	structBoost := 1.0
	if ctx.LearnStruct {
		if ctx.PrevSens != "" && ctx.LastSens != "" && ctx.PrevSens != ctx.LastSens {
			k := pairKey(ctx.PrevSens, ctx.LastSens)

			
			if v, ok := ctx.SeenPairs[k]; ok && v < 0 {
				
			} else {
				ctx.SeenPairs[k] += 0.40 * structBoost
				if ctx.SeenPairs[k] >= 1.0 {
					name := canonicalPairName(ctx.PrevSens, ctx.LastSens)
					id := "COACT:" + name

					if _, exists := ctx.Blocks[id]; !exists {
						ctx.AddBlock(NewCoActBlock(ctx.PrevSens, ctx.LastSens))
						ctx.TrainEvents = append(ctx.TrainEvents, fmt.Sprintf("+++ LEARNED NEW PAIR BLOCK %s", name))
					}

					ctx.SeenPairs[k] = -1.0
				}
			}
		}
	}

	
	if ctx.LearnStruct && !ctx.DisableSeq {
		if ctx.PrevSens != "" && ctx.LastSens != "" && ctx.PrevSens != ctx.LastSens {
			sk := ctx.PrevSens + ">" + ctx.LastSens

			if v, ok := ctx.SeenSeq[sk]; ok && v < 0 {
				
			} else {
				ctx.SeenSeq[sk] += 0.45 * structBoost
				if ctx.SeenSeq[sk] >= 1.0 {
					name := fmt.Sprintf("(%s>%s)", ctx.PrevSens, ctx.LastSens)
					id := "SEQ:" + name

					if _, exists := ctx.Blocks[id]; !exists {
						ctx.AddBlock(NewSeqBlock(ctx.PrevSens, ctx.LastSens))
						ctx.TrainEvents = append(ctx.TrainEvents, fmt.Sprintf("+++ LEARNED NEW SEQ BLOCK %s", name))

						actName := "ACT_ON_" + name
						ctx.AddBlock(NewActionBlock(name, actName))
						ctx.TrainEvents = append(ctx.TrainEvents, fmt.Sprintf("+++ ATTACHED ACTION %s <- %s", actName, name))
					}

					ctx.SeenSeq[sk] = -1.0
				}
			}
		}
	}

	
	if ctx.LearnStruct {
		if ctx.LastSens != "" && len(ctx.PrevStructSet) > 0 {
			for base := range ctx.PrevStructSet {
				a, b, ok := parsePairMembers(base)
				if !ok {
					continue
				}
				if ctx.LastSens == a || ctx.LastSens == b {
					continue
				}

				ck := base + "||" + ctx.LastSens
				if v, ok := ctx.SeenComposes[ck]; ok && v < 0 {
					continue
				}

				ctx.SeenComposes[ck] += 0.28 * structBoost
				if ctx.SeenComposes[ck] >= 1.0 {
					name := fmt.Sprintf("[%s-%s]", base, ctx.LastSens)
					id := "COMPOSE:" + name

					if _, exists := ctx.Blocks[id]; !exists {
						ctx.AddBlock(NewComposeBlock(base, ctx.LastSens))
						ctx.TrainEvents = append(ctx.TrainEvents, fmt.Sprintf("+++ LEARNED NEW COMPOSE BLOCK %s", name))

						actName := "ACT_ON_" + name
						ctx.AddBlock(NewActionBlock(name, actName))
						ctx.TrainEvents = append(ctx.TrainEvents, fmt.Sprintf("+++ ATTACHED ACTION %s <- %s", actName, name))
					}

					ctx.SeenComposes[ck] = -1.0
				}
			}
		}
	}

	
	if ctx.LearnPred {
		if ctx.LastSens != "" && len(ctx.PrevStructSet) > 0 {
			for st := range ctx.PrevStructSet {
				if _, ok := ctx.TransCounts[st]; !ok {
					ctx.TransCounts[st] = make(map[string]float64)
				}

				// base reinforcement (normal observation)
				learnRate := 0.22
				if ctx.ErrTTL > 0 {
					learnRate = 0.12
				}
				ctx.TransCounts[st][ctx.LastSens] += learnRate

				
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

				const (
					confirmN      = 4
					evidenceStep  = 0.22
					minMarginFrac = 0.35
				)
				eps := 1e-9

				computeGatedConf := func(targetTok string, targetV float64) float64 {
					if sumV <= eps || targetTok == "" || targetV <= 0 {
						return 0.0
					}

					rawConf := targetV / sumV

					
					minEvidenceForFull := float64(confirmN) * evidenceStep
					eGate := 1.0
					if minEvidenceForFull > eps {
						eGate = targetV / minEvidenceForFull
						if eGate > 1.0 {
							eGate = 1.0
						}
						if eGate < 0.0 {
							eGate = 0.0
						}
					}

					
					secondV := 0.0
					for tok, v := range ctx.TransCounts[st] {
						if tok == targetTok {
							continue
						}
						if v > secondV {
							secondV = v
						}
					}

					mGate := 1.0
					if secondV > 0 {
						need := secondV * (1.0 + minMarginFrac)
						if targetV < need {
							mGate = targetV / (need + eps)
							if mGate > 1.0 {
								mGate = 1.0
							}
							if mGate < 0.0 {
								mGate = 0.0
							}
						}
					}

					conf := rawConf * eGate * mGate

					// hard caps: no "absolute certainty" look
					if eGate < 1.0 || mGate < 1.0 {
						if conf > 0.95 {
							conf = 0.95
						}
					} else {
						if conf > 0.99 {
							conf = 0.99
						}
					}

					if conf < 0.01 {
						return 0.0
					}
					return conf
				}

				if bestTok != "" && sumV > 0 {
					oldV := 0.0
					if oldPred != "" {
						oldV = ctx.TransCounts[st][oldPred]
					}

					noPrev := oldPred == ""
					sameAsPrev := bestTok == oldPred

					
					domFactor := 1.35
					if ctx.ErrTTL > 0 {
						domFactor = 1.45 
					}
					canSwitchByStrength := (bestV >= PredSwitchMass) && (noPrev || bestV >= oldV*domFactor)

					
					pressureOverride := false
					if hadErrThisTick && !noPrev && !sameAsPrev {
						
						const (
							overrideRatio = 1.80 
							overrideAbs   = 0.60 
							overrideMass  = 1.20 
							minEvidence   = 0.88 
						)
						if bestV >= overrideMass && bestV >= minEvidence && bestV >= oldV*overrideRatio && (bestV-oldV) >= overrideAbs {
							pressureOverride = true
						}
					}

					allowSwitch := noPrev || sameAsPrev || canSwitchByStrength || pressureOverride

					if allowSwitch {
						ctx.BestPred[st] = bestTok
						ctx.PredConf[st] = computeGatedConf(bestTok, bestV)
					} else {
						
						ctx.BestPred[st] = oldPred
						if oldPred != "" {
							v := ctx.TransCounts[st][oldPred]
							c := computeGatedConf(oldPred, v)
							ctx.PredConf[st] = c * 0.92
						} else {
							ctx.PredConf[st] = oldConf * 0.92
						}
					}
				} else {
					
					ctx.BestPred[st] = oldPred
					ctx.PredConf[st] = oldConf * 0.92
					if ctx.PredConf[st] < 0.01 {
						ctx.PredConf[st] = 0.0
					}
				}

				if ctx.BestPred[st] != "" &&
					(ctx.BestPred[st] != oldPred || (ctx.PredConf[st]-oldConf) > 0.15) {
					if showPredEvents && !ctx.SuppressPredLog {
						ctx.PredEvents = append(
							ctx.PredEvents,
							fmt.Sprintf("+++ PREDICTION UPDATED: %s -> %s (conf=%.2f)", st, ctx.BestPred[st], ctx.PredConf[st]),
						)
					}
				}
			}
		}
	}
}



func resetEpisodeBoundary(ctx *Context) {
	
	ctx.PrevSens, ctx.LastSens = "", ""

	
	clearBoolMap(ctx.PrevStructSet)
	clearBoolMap(ctx.ThisStructSet)

		
	clearStringMap(ctx.ThisExpect)
	clearStringMap(ctx.PendingExpect)

	
	ctx.RecentActs = ctx.RecentActs[:0]
	ctx.RecentStruct = ctx.RecentStruct[:0]

	
	ctx.EnergySpentEpisode = 0
	ctx.ErrTTL = 0

	
	clearFloatMap(ctx.ThisStructMass)
	clearFloatMap(ctx.Inhib)
	clearIntMap(ctx.ErrCooldown)

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

func uniqueKeepOrder(xs []string) []string {
	seen := make(map[string]bool, len(xs))
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
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

func demoPulse(ctx *Context, title string) {
	

	errBoost := "OFF"
	if ctx.ErrTTL > 0 {
		errBoost = fmt.Sprintf("ON ttl=%d gain=%.2f", ctx.ErrTTL, ctx.ErrGain)
	}

	inhs := topInhibitions(ctx, 3)

	
	all := make([]string, 0, 16)
	for st, tok := range ctx.BestPred {
		if tok == "" {
			continue
		}
		conf := ctx.PredConf[st]
		if conf < 0.25 {
			continue
		}
		all = append(all, fmt.Sprintf("%s⇒%s(%.2f)", st, tok, conf))
	}
	sort.Strings(all)
	if len(all) > 3 {
		all = all[:3]
	}

	cprintf(
	    C_MAGENTA,
	    "DEMO-PULSE: %s | energy=%.2f spent=%.2f | errBoost=%s | inhib=%v | topExp=%v\n",
	    title, ctx.Energy, ctx.EnergySpentEpisode, errBoost, inhs, all,
	)

}

func armedExpectations(ctx *Context) []string {
	if ctx == nil || len(ctx.LastArmedExpect) == 0 {
		return nil
	}

	out := make([]string, 0, len(ctx.LastArmedExpect))
	for st, tok := range ctx.LastArmedExpect {
		if tok == "" {
			continue
		}
		conf := ctx.LastArmedConf[st]
		out = append(out, fmt.Sprintf("%s⇒%s(st=%.2f)", st, tok, conf))
	}
	sort.Strings(out)
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

func suppressedFromErrs(ctx *Context, episodeErrs []string, limit int) []string {
	// episodeErrs items look like: "(1>2):4->5" or "[1-2]:3->4"
	seen := make(map[string]bool)
	out := make([]string, 0, limit)

	for _, ev := range episodeErrs {
		parts := strings.SplitN(ev, ":", 2)
		if len(parts) != 2 {
			continue
		}
		st := parts[0]
		if st == "" || seen[st] {
			continue
		}
		seen[st] = true

		inh := ctx.Inhib[st]
		out = append(out, fmt.Sprintf("%s:%.2f", st, inh))
		if len(out) >= limit {
			break
		}
	}
	return out
}

func printBoard(ctx *Context, episodeStructs, episodeActions, episodeErrs, episodeTrain []string) {
	mode := "TRAIN"
	if !ctx.LearningEnabled {
		mode = "TEST"
	}

	pairs := countBlocksByPrefix(ctx, "COACT:")
	seqs := countBlocksByPrefix(ctx, "SEQ:")
	comps := countBlocksByPrefix(ctx, "COMPOSE:")
	acts := countBlocksByPrefix(ctx, "ACTIONBLOCK:")

	cprintf(C_MAGENTA+C_BOLD, "=== BOARD t=%03d mode=%s ===\n", ctx.Tick, mode)

	
	if ctx.DemoFocusPairsOnly {
	    cprintf(C_GREEN, "LEARNED: pairs=%d composes=%d actionLinks=%d blocks=%d\n", pairs, comps, acts, len(ctx.Blocks))
		} else {
		    cprintf(C_GREEN, "LEARNED: pairs=%d seqs=%d composes=%d actionLinks=%d blocks=%d\n", pairs, seqs, comps, acts, len(ctx.Blocks))
		}


	fmt.Printf("FIELD: energy=%.2f/%.2f\n", ctx.Energy, ctx.EnergyMax)
	fmt.Printf("FIELD: energy_spent_episode=%.2f\n", ctx.EnergySpentEpisode)

	
	if ctx.LastCleanupCount > 0 {
		recentWindow := ctx.PruneEvery
		if recentWindow <= 0 {
			recentWindow = 10
		}
		if ctx.Tick-ctx.LastCleanupTick >= 0 && ctx.Tick-ctx.LastCleanupTick <= recentWindow {
			fmt.Printf(
				"SELF-CLEANUP: pruned=%d inactive blocks (age>=%d ticks, every=%d ticks)\n",
				ctx.LastCleanupCount,
				ctx.ForgetAfter,
				ctx.PruneEvery,
			)
		}
	}

	lastPairs := blockNamesByPrefixLast(ctx, "COACT:", 5)
	lastSeqs := blockNamesByPrefixLast(ctx, "SEQ:", 5)
	lastComps := blockNamesByPrefixLast(ctx, "COMPOSE:", 5)

	if len(lastPairs) > 0 {
		fmt.Printf("LAST PAIRS:   %v\n", lastPairs)
	}
	if len(lastSeqs) > 0 && !ctx.DemoFocusPairsOnly {
		fmt.Printf("LAST SEQ:     %v\n", lastSeqs)
	}
	if len(lastComps) > 0 {
		fmt.Printf("LAST COMPOSE: %v\n", lastComps)
	}

	// Episode summary
	es := uniqueSorted(episodeStructs)
	ea := uniqueSorted(episodeActions)

	// IMPORTANT: keep order for errors/training
	ee := uniqueKeepOrder(episodeErrs)
	et := uniqueKeepOrder(episodeTrain)

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
		cprintf(C_RED+C_BOLD, "EPISODE: errors=%v\n", derrs)
	}

	
	if len(et) > 0 {
		learned := make([]string, 0, len(et))
		attached := make([]string, 0, len(et))
		other := make([]string, 0, len(et))

		for _, ev := range et {
			switch {
			case strings.Contains(ev, "+++ LEARNED NEW"):
				learned = append(learned, ev)
			case strings.Contains(ev, "+++ ATTACHED ACTION"):
				attached = append(attached, ev)
			default:
				other = append(other, ev)
			}
		}

		raw := append(append(learned, attached...), other...)
		dtrain := make([]string, 0, len(raw))
		for _, ev := range raw {
			dtrain = append(dtrain, strings.ReplaceAll(ev, "->", "⇒"))
		}
		if len(dtrain) > 6 {
			dtrain = dtrain[:6]
		}

		fmt.Printf("TRAINING: events=%d  sample=%v\n", len(et), dtrain)
	}

	
	armed := armedExpectations(ctx)
	if len(armed) == 0 {
		fmt.Println("FIELD: armed expectations=(none)")
	} else {
		fmt.Printf("FIELD: armed expectations=%v\n", armed)
	}

	
	preds := topPredictions(ctx, episodeStructs, 6)
	if len(preds) == 0 {
		fmt.Println("FIELD: model expectations=(none)")
	} else {
		fmt.Printf("FIELD: model expectations=%v\n", preds)
	}

	
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
		if len(all) > 8 {
			all = all[:8]
		}
		
		if strings.Join(all, "|") != strings.Join(armed, "|") {
			fmt.Printf("FIELD: all expectations=%v\n", all)
		}
	}

	
	inhs := topInhibitions(ctx, 6)
	if len(inhs) == 0 {
		fmt.Println("FIELD: inhib=(none)")
	} else {
		fmt.Printf("FIELD: inhib=%v\n", inhs)
	}

	
	supp := suppressedFromErrs(ctx, ee, 6)
	if len(supp) > 0 {
		fmt.Printf("FIELD: suppressed=%v\n", supp)
	}

	
	if len(ee) > 0 && len(ctx.LastAdapt) > 0 {
		ad := ctx.LastAdapt
		if len(ad) > 4 {
			ad = ad[:4]
		}
		cprintf(C_GREEN+C_BOLD, "ADAPTATION: %v\n", ad)

	}

	
	
	if len(ee) > 0 {
		cprintf(C_YELLOW+C_BOLD, "LEARNING: error-boost=ON (episode had ERR) gain=%.2f\n", ctx.ErrGain)
	} else if ctx.ErrTTL > 0 {
		cprintf(C_YELLOW+C_BOLD, "LEARNING: error-boost=ON ttl=%d gain=%.2f\n", ctx.ErrTTL, ctx.ErrGain)
	} else {
		cprintf(C_GRAY, "LEARNING: error-boost=OFF\n")
	}
}

func parseErrTriplet(ev string) (st, pred, actual string, ok bool) {
	
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

type EpisodeReport struct {
	Structs []string
	Actions []string
	Errs    []string
}

func RunEpisodeLine(ctx *Context, line string, investorMode bool, demoRunning bool, sleepMs int, autoBoard bool) EpisodeReport {
	resetEpisodeBoundary(ctx)
	ctx.LastAdapt = ctx.LastAdapt[:0]
	tokens := strings.Fields(strings.TrimSpace(line))
	return RunEpisodeTokens(ctx, tokens, investorMode, demoRunning, sleepMs, autoBoard)
}


func RunEpisodeTokens(ctx *Context, tokens []string, investorMode bool, demoRunning bool, sleepMs int, autoBoard bool) EpisodeReport {
	episodeStructs := make([]string, 0, 32)
	episodeActions := make([]string, 0, 32)
	episodeErrs := make([]string, 0, 32)
	episodePredEvents := make([]string, 0, 32)
	episodeTrainEvents := make([]string, 0, 32)

	
	mispShown := 0
	const mispLimit = 2
	mispDropped := 0
	mispSummaryPrinted := false

	for i, tok := range tokens {
		
		if !ctx.Sensors[tok] {
			ctx.AddBlock(&SensorBlock{token: tok})
			ctx.Sensors[tok] = true

			if ctx.LearningEnabled {
				fmt.Printf("+++ AUTO-SENSOR CREATED [%s]\n", tok)
			} else {
				fmt.Printf("+++ TOKEN REGISTERED [%s] (test mode; no learning)\n", tok)
			}
		}

		inSig := Signal{Kind: K_SENS, Value: tok, Mass: 1.0, Time: ctx.Tick, From: "USER"}

		
		oldConf := make(map[string]float64, len(ctx.PendingExpect))
		oldBest := make(map[string]string, len(ctx.PendingExpect))
		for st := range ctx.PendingExpect {
			oldConf[st] = ctx.PredConf[st]
			oldBest[st] = ctx.BestPred[st]
		}

		
		oldLast := ctx.LastSens

		oldPair := 0.0
		oldSeq := 0.0
		pairName := ""
		seqName := ""
		pairK := ""
		seqK := ""

		if oldLast != "" && oldLast != tok {
			pairK = pairKey(oldLast, tok)
			oldPair = ctx.SeenPairs[pairK]
			pairName = canonicalPairName(oldLast, tok)

			seqK = oldLast + ">" + tok
			oldSeq = ctx.SeenSeq[seqK]
			seqName = fmt.Sprintf("(%s>%s)", oldLast, tok)
		}

		
		oldCompose := make(map[string]float64, 4)
		composeName := make(map[string]string, 4)
		if len(ctx.PrevStructSet) > 0 {
			for base := range ctx.PrevStructSet {
				a, b, ok := parsePairMembers(base)
				if !ok {
					continue
				}
				if tok == a || tok == b {
					continue
				}
				ck := base + "||" + tok
				oldCompose[ck] = ctx.SeenComposes[ck]
				composeName[ck] = fmt.Sprintf("[%s-%s]", base, tok)
			}
		}

		
		oldTrans := make(map[string]float64, 4)
		if ctx.ErrTTL == 0 && len(ctx.PrevStructSet) > 0 {
			for st := range ctx.PrevStructSet {
				if m, ok := ctx.TransCounts[st]; ok {
					oldTrans[st] = m[tok]
				} else {
					oldTrans[st] = 0
				}
			}
		}

		ctx.PredEvents = ctx.PredEvents[:0]
		ctx.TrainEvents = ctx.TrainEvents[:0]

		out := RunTick(ctx, []Signal{inSig})

		
		predEvents := append([]string(nil), ctx.PredEvents...)
		trainEvents := append([]string(nil), ctx.TrainEvents...)

		
		if len(predEvents) > 0 {
			episodePredEvents = append(episodePredEvents, predEvents...)
		}
		if len(trainEvents) > 0 {
			episodeTrainEvents = append(episodeTrainEvents, trainEvents...)
		}

		actions := make([]string, 0, 8)
		structs := make([]string, 0, 8)
		errs := make([]string, 0, 8)

		for _, s := range out {
			if s.Kind == K_ACTION {
				actions = append(actions, s.Value)
				episodeActions = append(episodeActions, s.Value)
			}

			if s.Kind == K_STRUCT {
				
				if ctx.DemoFocusPairsOnly {
					isPair := strings.HasPrefix(s.Value, "[") && strings.HasSuffix(s.Value, "]")
					if !isPair {
						
						goto SKIP_STRUCT_APPEND
					}
				}

				structs = append(structs, s.Value)
				episodeStructs = append(episodeStructs, s.Value)

			SKIP_STRUCT_APPEND:
				
				_ = 0
			}

			if s.Kind == K_ERR {
				errs = append(errs, s.Value)
				episodeErrs = append(episodeErrs, s.Value)
			}
		}

		hadErrThisTick := len(errs) > 0

		
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
			
			if !(ctx.DemoFocusPairsOnly && demoRunning) {
				now := ctx.SeenSeq[seqK]
				if now > oldSeq {
					chargeLines = append(chargeLines,
						fmt.Sprintf("CHARGE SEQ  %s  mass=%.2f/1.00 (+%.2f)", seqName, now, now-oldSeq),
					)
					if now >= 0.85 && now < 1.0 {
						chargeLines = append(chargeLines, fmt.Sprintf("NEAR-CRYSTAL %s", seqName))
					}
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
					chargeLines = append(chargeLines, fmt.Sprintf("NEAR-CRYSTAL %s", nm))
				}
			}
		}

		
		if !hadErrThisTick {
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
		}

		
		tickHadEvent := false

		
		if len(structs) > 0 || len(actions) > 0 || len(errs) > 0 || len(predEvents) > 0 || len(trainEvents) > 0 {
			tickHadEvent = true
		}

		
		if len(chargeLines) > 0 {
			for _, ln := range chargeLines {
				if strings.Contains(ln, "NEAR-CRYSTAL") {
					tickHadEvent = true
					break
				}
			}
		}

		
		if investorMode {
			showWarmup := demoRunning && i < 2
			showLast := demoRunning && i == len(tokens)-1

			
			if !showWarmup && !showLast &&
				len(structs) == 0 && len(actions) == 0 && len(errs) == 0 &&
				len(chargeLines) == 0 && len(predEvents) == 0 && len(trainEvents) == 0 {

				
				cprintf(C_GRAY, "t=%03d INPUT=%s\n", ctx.Tick, tok)


				if sleepMs > 0 {
					time.Sleep(time.Duration(sleepMs) * time.Millisecond)
				}
				continue
			}

			
			if !showWarmup &&
				len(structs) == 0 && len(actions) == 0 && len(errs) == 0 &&
				len(chargeLines) > 0 && len(predEvents) == 0 && len(trainEvents) == 0 {

				cprintf(C_GRAY, "t=%03d INPUT=%s\n", ctx.Tick, tok)

				for _, ln := range chargeLines {
					fmt.Printf("           %s\n", ln)
				}

				
				if tickHadEvent {
					fmt.Printf("           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)
				}

				if sleepMs > 0 {
					time.Sleep(time.Duration(sleepMs) * time.Millisecond)
				}
				continue
			}
		}

		if len(structs) > 0 {
			cprintf(C_GRAY, "t=%03d INPUT=%s  ", ctx.Tick, tok)
			cprintf(C_CYAN+C_BOLD, "STRUCT=%v\n", structs)
		} else {
			cprintf(C_GRAY, "t=%03d INPUT=%s\n", ctx.Tick, tok)
		}


		
		if !investorMode && len(chargeLines) > 0 {
			for _, ln := range chargeLines {
				color := C_BLUE
				if strings.Contains(ln, "NEAR-CRYSTAL") {
					color = C_YELLOW + C_BOLD
				}
				cprintf(color, "           %s\n", ln)
			}

		}

		
		if len(actions) > 0 && !(investorMode && demoRunning) {
			cprintf(C_GREEN+C_BOLD, "           ACTION=%v\n", actions)
		}
		if len(errs) > 0 {
			cprintf(C_RED+C_BOLD, "           ERROR=%v\n", errs)
		}

		
		suppressed := make(map[string]bool, 8)

		
		allowVerboseMisp := !investorMode || len(errs) == 0 || (mispShown < mispLimit)

		if investorMode && len(errs) > 0 && allowVerboseMisp {
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
				parts = append(parts, fmt.Sprintf("%s⇒%s(conf=%.2f)", st, showPred, conf))
			}

			if len(parts) > 0 {
				cprintf(C_MAGENTA, "           EXPECTATIONS: %s\n", strings.Join(parts, " ; "))
			}
		}

		if len(errs) > 0 {
			if allowVerboseMisp {
				for _, ev := range errs {
					st, pred, actual, ok := parseErrTriplet(ev)
					if !ok {
						continue
					}
					if ctx.Inhib[st] > 0.0 {
						suppressed[st] = true
					}

					showPred := pred
					if b := oldBest[st]; b != "" {
						showPred = b
					}

					before := oldConf[st]

					
					instantAfter := before * 0.70
					if instantAfter < 0 {
						instantAfter = 0
					}

					if ctx.LearningEnabled {
						cprintf(
							C_RED,
							"           MISPREDICTION: %s expected %s (conf %.2f) ⇒ got %s (conf %.2f->%.2f)\n",
							st, showPred, before, actual, before, instantAfter,
						)

					} else {
						cprintf(
							C_RED,
							"           MISPREDICTION: %s expected %s (conf %.2f) ⇒ got %s\n",
							st, showPred, before, actual,
						)

						cprintf(C_YELLOW, "           NOTE: TEST mode => no learning; switch to TRAIN to adapt\n")
					}
				}

				if investorMode && len(errs) > 0 {
					inhibN := len(suppressed)
					if inhibN == 0 {
						inhibN = len(errs)
					}
					cprintf(C_YELLOW+C_BOLD,
						"           FIELD RESPONSE: inhibited=%d | error-boost ttl=%d gain=%.2f\n",
						inhibN, ctx.ErrTTL, ctx.ErrGain,
					)

				}

				
				if investorMode {
					mispShown++
				}
			} else {
				
				if investorMode {
					mispDropped++
				}
			}
		}

		
		if tickHadEvent || (demoRunning && i == len(tokens)-1) {
			cprintf(C_GRAY, "           ENERGY_NOW=%.2f  SPENT_EP=%.2f\n", ctx.Energy, ctx.EnergySpentEpisode)
		}

		
		if !demoRunning && len(trainEvents) > 0 {
			for _, te := range trainEvents {
				cprintf(C_GREEN, "           %s\n", te)
			}
		}

		
		if len(predEvents) > 0 {
			for _, pe := range predEvents {
				cprintf(C_CYAN, "           %s\n", pe)
			}
		}

		if sleepMs > 0 {
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		}

		
		if autoBoard && i == len(tokens)-1 {
			episodeMeaningful :=
				len(episodeStructs) > 0 ||
					len(episodeActions) > 0 ||
					len(episodeErrs) > 0 ||
					len(episodePredEvents) > 0 ||
					len(episodeTrainEvents) > 0 ||
					ctx.EnergySpentEpisode > 0

			if episodeMeaningful {
				
				if investorMode && mispDropped > 0 && !mispSummaryPrinted {
					fmt.Printf("MISPREDICTION: (+%d more suppressed for readability)\n", mispDropped)
					mispSummaryPrinted = true
				}
				printBoard(ctx, episodeStructs, episodeActions, episodeErrs, episodeTrainEvents)
			}
		}
	}

	
	if investorMode && mispDropped > 0 && !mispSummaryPrinted {
		fmt.Printf("MISPREDICTION: (+%d more suppressed for readability)\n", mispDropped)
		mispSummaryPrinted = true
	}

	return EpisodeReport{Structs: episodeStructs, Actions: episodeActions, Errs: episodeErrs}
}




func main() {
	ctx := NewContext()

	
	sleepMs := 12 
	autoBoard := true
	investorMode := false 

	
	demoRunning := false

	fmt.Println("STB DEMO (INHIB+PRED+ERROR+FORGET): signals -> blocks -> competition -> prediction -> error-driven learning -> forgetting.")
	fmt.Println("Commands: train | test | reset | board | demo | quit")
	fmt.Println("Suggested demo:")
	fmt.Println("  demo   (runs 3 steps:")
	fmt.Println("          1) crystallize pairs [1-2] and [2-3]")
	fmt.Println("          2) show stable prediction [1-2]⇒3")
	fmt.Println("          3) clean misprediction 3⇒4 with error-boost and fast re-learn)")
	fmt.Println("  manual alternative (same logic, explicit episode boundaries):")
	fmt.Println("    train ; reset ; repeat: 1 2 1 2 1 2   2 3 2 3 2 3   (crystallization)")
	fmt.Println("    train ; reset ; repeat: 1 2 3 1 2 3 1 2 3           (stable 1-2⇒3)")
	fmt.Println("    train ; reset ; run:    1 2 3                         (prime expectation)")
	fmt.Println("    train ; reset ; run:    1 2 4                         (clean 3⇒4 switch)")
	fmt.Println("    train ; reset ; run:    1 2 4                         (verify adaptation)")
	fmt.Println("Input tokens separated by spaces. Example: 1 2 1 2 1 2 3 1 2 3 1 2 4")

	in := bufio.NewScanner(os.Stdin)

	
	lastEpisodeStructs := []string{}
	lastEpisodeActions := []string{}
	lastEpisodeErrs := []string{}

	
	lastBoardCtx := ctx

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
	    ctx.LearnStruct = true
	    ctx.LearnPred = true
	    fmt.Println("MODE = TRAIN (learning enabled)")
	    continue

		case "test":
	    ctx.LearningEnabled = false
	    ctx.LearnStruct = false
	    ctx.LearnPred = false
	    fmt.Println("MODE = TEST (learning disabled)")
	    continue

	    case "color on":
			colorLogs = true
			fmt.Println("Color logs = ON")
			continue

		case "color off":
			colorLogs = false
			fmt.Println("Color logs = OFF")
			continue

		case "reset":
			resetEpisodeBoundary(ctx)
			
			lastEpisodeStructs = nil
			lastEpisodeActions = nil
			lastEpisodeErrs = nil
			fmt.Println("Reset episode boundary")
			continue

		case "board":
			printBoard(lastBoardCtx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs, nil)
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

		case "pairs on":
		    ctx.DemoFocusPairsOnly = true
		    fmt.Println("Pairs-only mode = ON (UI-only: hides non-[a-b] in episode/board output)")
		    continue

		case "pairs off":
		    ctx.DemoFocusPairsOnly = false
		    fmt.Println("Pairs-only mode = OFF (UI-only)")
		    continue



			
case "demo":
	
	prevInvestorMode := investorMode
	prevAutoBoard := autoBoard
	prevSleepMs := sleepMs
	prevDemoRunning := demoRunning

	
	demoRunning = true
	investorMode = true
	autoBoard = false
	sleepMs = 0
	fmt.Println("Demo: investor mode ON, running scripted sequence...")

	
	demoCtx := NewContext()
	demoCtx.LearningEnabled = true
	demoCtx.DemoFocusPairsOnly = true
	demoCtx.DisableSeq = true

	
	lastEpisodeStructs = nil
	lastEpisodeActions = nil
	lastEpisodeErrs = nil

	
	demoCtx.SuppressPredLog = true
	demoCtx.LearnStruct = true
	demoCtx.LearnPred = false

	fmt.Println("DEMO STEP 1/3: ACCUMULATION -> CRYSTALLIZATION")

	
	rep := RunEpisodeLine(
		demoCtx,
		"1 2 1 2 1 2 1 2 1 2 1 2   2 3 2 3 2 3 2 3 2 3 2 3",
		investorMode, demoRunning, sleepMs, false,
	)

	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs

	printBoard(demoCtx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs, nil)
	lastBoardCtx = demoCtx

	fmt.Println("NOTE: Step 1 reports accumulation and block crystallization (new blocks). STRUCT signals appear in Step 2.")

	
	demoCtx.SuppressPredLog = false
	demoCtx.LearnStruct = false
	demoCtx.LearnPred = true
	demoCtx.DemoFocusPairsOnly = true

	fmt.Println("DEMO STEP 2/3: STRUCTURES -> PREDICTION")

	
	rep = RunEpisodeLine(
		demoCtx,
		"1 2 3 1 2 3 1 2 3 1 2 3 1 2 3 1 2 3",
		investorMode, demoRunning, sleepMs, false,
	)

	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs

	printBoard(demoCtx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs, nil)
	lastBoardCtx = demoCtx

	
	fmt.Println("DEMO STEP 3/3: MISPREDICTION -> INHIBITION + ERROR-BOOST -> FAST RE-LEARN")

	
	demoCtx.LearningEnabled = false
	demoCtx.LearnStruct = false
	demoCtx.LearnPred = false

	rep = RunEpisodeLine(demoCtx, "1 2 3", investorMode, demoRunning, sleepMs, false)
	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs
	demoPulse(demoCtx, "after prime episode: 1 2 3")

	
	demoCtx.LearningEnabled = true
	demoCtx.LearnStruct = false
	demoCtx.LearnPred = true

	rep = RunEpisodeLine(
		demoCtx,
		"1 2 4 1 2 4 1 2 4 1 2 4 1 2 4 1 2 4",
		investorMode, demoRunning, sleepMs, false,
	)

	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs
	demoPulse(demoCtx, "after clean switch episode: 1 2 4")
	printBoard(demoCtx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs, nil)

	
	demoCtx.LearningEnabled = true
	demoCtx.LearnStruct = false
	demoCtx.LearnPred = true

	rep = RunEpisodeLine(demoCtx, "1 2 4", investorMode, demoRunning, sleepMs, false)
	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs
	demoPulse(demoCtx, "verify #1 (train): 1 2 4")

	
	demoCtx.LearningEnabled = false
	demoCtx.LearnStruct = false
	demoCtx.LearnPred = false

	rep = RunEpisodeLine(demoCtx, "1 2 4", investorMode, demoRunning, sleepMs, false)
	lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs =
		rep.Structs, rep.Actions, rep.Errs

	demoPulse(demoCtx, "verify #2 (test): 1 2 4")
	printBoard(demoCtx, lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs, nil)
	lastBoardCtx = demoCtx

	
	pairs := countBlocksByPrefix(demoCtx, "COACT:")
	seqs := countBlocksByPrefix(demoCtx, "SEQ:")
	comps := countBlocksByPrefix(demoCtx, "COMPOSE:")
	acts := countBlocksByPrefix(demoCtx, "ACTIONBLOCK:")

	fmt.Printf(
		"DEMO SUMMARY: learned pairs=%d | seqs=%d | composes=%d | actionLinks=%d | blocks=%d\n",
		pairs, seqs, comps, acts, len(demoCtx.Blocks),
	)

	demoCtx.DemoFocusPairsOnly = false

	
	investorMode = prevInvestorMode
	autoBoard = prevAutoBoard
	sleepMs = prevSleepMs
	demoRunning = prevDemoRunning

	continue


		}

		
		rep := RunEpisodeLine(ctx, line, investorMode, demoRunning, sleepMs, autoBoard)
		lastEpisodeStructs, lastEpisodeActions, lastEpisodeErrs = rep.Structs, rep.Actions, rep.Errs
		lastBoardCtx = ctx
	}
}
