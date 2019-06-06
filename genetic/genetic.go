package genetic

import (
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Individual interface {
	Fitness() int
	Clone() Individual
	Shell() Individual
}

type LocalOptimization interface {
	LocallyOptimize() Individual
}

type LockedRand struct {
	G  Rand
	mu sync.Mutex
}

func (l *LockedRand) Intn(n int) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.G.Intn(n)
}

func (l *LockedRand) Int() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.G.Int()
}

var _ Rand = &LockedRand{}

type Simplifyable interface {
	Simplify()
}

type Array interface {
	Individual
	Swap(i, j int)
	Copy(from Array, start int, end int, into int)
	Len() int
}

type ExecutionTerminator interface {
	StopExecution(p Population) bool
}

type CountingExecutor struct {
	Limit int
	i     int
}

func (c *CountingExecutor) StopExecution(p Population) bool {
	if c.i >= c.Limit {
		return true
	}
	c.i++
	return false
}

var _ ExecutionTerminator = &CountingExecutor{}

type TimingExecutor struct {
	Duration  time.Duration
	startTime time.Time
	Now       func() time.Time
}

func (c *TimingExecutor) now() time.Time {
	if c.Now == nil {
		return time.Now()
	}
	return c.Now()
}

func (c *TimingExecutor) StopExecution(p Population) bool {
	if c.startTime.IsZero() {
		c.startTime = c.now()
	}
	curTime := c.now()
	return curTime.Sub(c.startTime) > c.Duration
}

var _ ExecutionTerminator = &TimingExecutor{}

type IndividualFactory interface {
	Spawn() Individual
}

type Mutator interface {
	Mutate(in Individual) Individual
}

type Breeder interface {
	Reproduce(in []Individual) Individual
}

type SwapMutator struct {
	MutationRatio int
	R             Rand
}

func (a *SwapMutator) Mutate(in Individual) Individual {
	if a.MutationRatio > 1 {
		if a.R.Intn(a.MutationRatio) != 0 {
			return in
		}
	}
	asSwap, ok := in.(Array)
	if !ok {
		panic("Trying to do swap mutation with something that isn't swappable")
	}

	i := a.R.Intn(asSwap.Len())
	j := a.R.Intn(asSwap.Len())
	if i == j {
		return in
	}
	asS := asSwap.Clone()
	asS.(Array).Swap(i, j)
	return asS
}

type DynamicMutation interface {
	IncreaseMutationRate()
	ResetMutationRate()
}

var _ Mutator = &SwapMutator{}

type LookAheadMutator struct {
	MutationRatio         int
	currentMutationRation int
	R                     Rand
}

func (a *LookAheadMutator) IncreaseMutationRate() {
	a.currentMutationRation--
	if a.currentMutationRation <= 1 {
		a.currentMutationRation = 1
	}
}

func (a *LookAheadMutator) ResetMutationRate() {
	a.currentMutationRation = a.MutationRatio
}

func (a *LookAheadMutator) Mutate(in Individual) Individual {
	if a.R.Intn(a.currentMutationRation) != 0 {
		return in
	}
	asSwap, ok := in.(Array)
	if !ok {
		panic("Trying to do swap mutation with something that isn't swappable")
	}
	newPerson := asSwap.Clone().(Array)
	startingIndex := a.R.Intn(asSwap.Len())
	for i := 0; i < asSwap.Len(); i++ {
		shouldSwap := a.R.Intn(asSwap.Len()) == 0
		if !shouldSwap {
			continue
		}
		swapWith := a.R.Intn(asSwap.Len())
		if swapWith == i {
			continue
		}
		newPerson.Swap((i+startingIndex)%asSwap.Len(), (swapWith+startingIndex)%asSwap.Len())
	}
	return newPerson
}

var _ Mutator = &LookAheadMutator{}
var _ DynamicMutation = &LookAheadMutator{}

type Rand interface {
	Intn(int) int
	Int() int
}

var _ Rand = &rand.Rand{}

type SplitReproduce struct {
	R Rand
}

func (s *SplitReproduce) Reproduce(in []Individual) Individual {
	if len(in) == 0 {
		return nil
	}
	if len(in) == 1 {
		return in[0]
	}
	if len(in) > 2 {
		panic("I haven't implemented >2 yet")
	}
	asSwap, ok := in[0].(Array)
	if !ok {
		panic("split reproducer only allowed on swappable")
	}
	midPoint := s.R.Intn(asSwap.Len())
	prefC := s.R.Intn(2)%2 == 0
	ret := in[0].Shell().(Array)
	if prefC {
		ret.Copy(in[0].(Array), 0, midPoint, 0)
		ret.Copy(in[1].(Array), midPoint, in[1].(Array).Len(), midPoint)
	} else {
		ret.Copy(in[1].(Array), 0, midPoint, 0)
		ret.Copy(in[0].(Array), midPoint, in[1].(Array).Len(), midPoint)
	}
	return ret
}

var _ Breeder = &SplitReproduce{}

type ParentSelector interface {
	PickParent([]Individual) int
}

type TournamentParentSelector struct {
	R Rand
	K int
}

func (s TournamentParentSelector) PickParent(c []Individual) int {
	k := s.K
	if k == 0 {
		k = int(math.Log(float64(len(c))) + 1)
	}
	current := s.R.Intn(len(c))
	for i := 1; i < k; i++ {
		other := s.R.Intn(len(c))
		if c[current].Fitness() < c[other].Fitness() {
			current = other
		}
	}
	return current
}

var _ ParentSelector = &TournamentParentSelector{}

type Population struct {
	People   []Individual
	isSorted bool
}

func SpawnPopulation(n int, f IndividualFactory) Population {
	ret := Population{
		People: make([]Individual, n),
	}
	for i := range ret.People {
		ret.People[i] = f.Spawn()
	}
	return ret
}

func (p *Population) Sort() {
	if p.isSorted {
		return
	}
	sort.Slice(p.People, func(i, j int) bool {
		return p.People[i].Fitness() < p.People[j].Fitness()
	})
	p.isSorted = true
}

func (p *Population) Min() Individual {
	worst := p.People[0]
	for i := 1; i < len(p.People); i++ {
		if p.People[i].Fitness() < worst.Fitness() {
			worst = p.People[i]
		}
	}
	return worst
}

func (p *Population) Max() Individual {
	best := p.People[0]
	for i := 1; i < len(p.People); i++ {
		if p.People[i].Fitness() > best.Fitness() {
			best = p.People[i]
		}
	}
	return best
}

func (p *Population) Average() float64 {
	sum := 0
	for _, c := range p.People {
		sum += c.Fitness()
	}
	return float64(sum) / float64(len(p.People))
}

func (p *Population) calculateFitness(numGoroutine int) {
	if numGoroutine < 2 {
		for i := 0; i < len(p.People); i++ {
			p.People[i].Fitness()
		}
		return
	}
	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	idxChan := make(chan int)
	for i := 0; i < numGoroutine; i++ {
		go func() {
			defer wg.Done()
			for idx := range idxChan {
				p.People[idx].Fitness()
			}
		}()
	}
	for i := 0; i < len(p.People); i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
}

func (p *Population) singleNextGenerationIteration(ps ParentSelector, b Breeder, m Mutator, numP int) Individual {
	parents := make([]Individual, numP)
	for j := 0; j < numP; j++ {
		parents[j] = p.People[ps.PickParent(p.People)]
	}
	newChild := b.Reproduce(parents)
	mutatedChild := m.Mutate(newChild)
	return mutatedChild
}

func (p *Population) NextGeneration(ps ParentSelector, b Breeder, m Mutator, numP int, numGoroutine int) Population {
	p.calculateFitness(numGoroutine)
	ret := Population{
		People: make([]Individual, len(p.People)),
	}
	if numGoroutine < 2 {
		numGoroutine = 1
	}
	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	idxChan := make(chan int)
	for i := 0; i < numGoroutine; i++ {
		go func() {
			defer wg.Done()
			for idx := range idxChan {
				ret.People[idx] = p.singleNextGenerationIteration(ps, b, m, numP)
			}
		}()
	}
	for i := 0; i < len(p.People)-1; i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
	ret.People[len(ret.People)-1] = m.Mutate(p.Max())
	return ret
}

type Algorithm struct {
	Log             *log.Logger
	ParentSelector  ParentSelector
	Factory         IndividualFactory
	Terminator      ExecutionTerminator
	Breeder         Breeder
	Mutator         Mutator
	NumberOfParents int
	PopulationSize  int
	NumGoroutine    int
}

func (a *Algorithm) Run() Individual {
	currentPopulation := SpawnPopulation(a.PopulationSize, a.Factory)
	best := currentPopulation.Max()
	asDynamic, isDynamic := a.Mutator.(DynamicMutation)
	if isDynamic {
		asDynamic.ResetMutationRate()
	}
	for {
		if a.Log != nil {
			a.Log.Println("Currently at mean/max", currentPopulation.Average(), currentPopulation.Max().Fitness())
		}
		if a.Terminator.StopExecution(currentPopulation) {
			return best
		}
		nextPopulation := currentPopulation.NextGeneration(a.ParentSelector, a.Breeder, a.Mutator, a.NumberOfParents, a.NumGoroutine)
		nextBest := nextPopulation.Max()
		if best.Fitness() < nextBest.Fitness() {
			best = nextPopulation.Max()
			if isDynamic {
				asDynamic.ResetMutationRate()
			}
		} else if isDynamic {
			asDynamic.IncreaseMutationRate()
		}
		currentPopulation = nextPopulation
	}
}
