package genetic

import (
	"sort"
	"sync"
)

type Population struct {
	People   []Individual
	isSorted bool
}

func SpawnPopulation(n int, f IndividualFactory, r Rand) Population {
	ret := Population{
		People: make([]Individual, n),
	}
	for i := range ret.People {
		ret.People[i] = f.Spawn(r)
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

func (p *Population) singleNextGenerationIteration(ps ParentSelector, b Breeder, m Mutator, numP int, rnd Rand) Individual {
	parents := make([]Individual, numP)
	for j := 0; j < numP; j++ {
		parents[j] = p.People[ps.PickParent(p.People, rnd)]
	}
	newChild := b.Reproduce(parents, rnd)
	mutatedChild := m.Mutate(newChild, rnd)
	return mutatedChild
}

func (p *Population) NextGeneration(ps ParentSelector, b Breeder, m Mutator, numP int, numGoroutine int, rnd RandForIndex) Population {
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
				ret.People[idx] = p.singleNextGenerationIteration(ps, b, m, numP, rnd.Rand(idx))
			}
		}()
	}
	for i := 0; i < len(p.People)-1; i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
	ret.People[len(ret.People)-1] = m.Mutate(p.Max(), rnd.Rand(0))
	return ret
}