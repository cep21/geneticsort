package genetic

import (
	"math/rand"
	"sync"
)

type RandForIndex interface {
	Rand(idx int) Rand
}

func ArrayRandForIdx(size int, seed int64, generator func(seed int64) Rand) RandForIndex {
	start := generator(seed)
	ret := &arrayRandForIdx{
		rands: make([]Rand, size),
	}
	for i := 0; i < size; i++ {
		ret.rands[i] = generator(start.Int63())
	}
	return ret
}

type arrayRandForIdx struct {
	rands []Rand
}

func (a *arrayRandForIdx) Rand(idx int) Rand {
	return a.rands[idx]
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

func (l *LockedRand) Int63() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.G.Int63()
}


type Rand interface {
	Intn(int) int
	Int() int
	Int63() int64
}

var _ Rand = &rand.Rand{}
var _ Rand = &LockedRand{}
