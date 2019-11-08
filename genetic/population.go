package genetic

import (
	"sort"
	"sync"
)

type Population struct {
	Individuals []Chromosome
	isSorted    bool
}

func SpawnPopulation(n int, f ChromosomeFactory, r Rand) Population {
	ret := Population{
		Individuals: make([]Chromosome, n),
	}
	for i := range ret.Individuals {
		ret.Individuals[i] = f.Spawn(r)
	}
	return ret
}

func (p *Population) Sort() {
	if p.isSorted {
		return
	}
	sort.Slice(p.Individuals, func(i, j int) bool {
		return p.Individuals[i].Fitness() < p.Individuals[j].Fitness()
	})
	p.isSorted = true
}

func (p *Population) Min() Chromosome {
	worst := p.Individuals[0]
	for i := 1; i < len(p.Individuals); i++ {
		if p.Individuals[i].Fitness() < worst.Fitness() {
			worst = p.Individuals[i]
		}
	}
	return worst
}

func (p *Population) Max() Chromosome {
	best := p.Individuals[0]
	for i := 1; i < len(p.Individuals); i++ {
		if p.Individuals[i].Fitness() > best.Fitness() {
			best = p.Individuals[i]
		}
	}
	return best
}

func (p *Population) Average() float64 {
	sum := 0
	for _, c := range p.Individuals {
		sum += c.Fitness()
	}
	return float64(sum) / float64(len(p.Individuals))
}

func (p *Population) calculateFitness(numGoroutine int) {
	if numGoroutine < 2 {
		for i := 0; i < len(p.Individuals); i++ {
			p.Individuals[i].Fitness()
		}
		return
	}
	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	individuals := make(chan Chromosome)
	for i := 0; i < numGoroutine; i++ {
		go func() {
			defer wg.Done()
			for individual := range individuals {
				individual.Fitness()
			}
		}()
	}
	for i := 0; i < len(p.Individuals); i++ {
		individuals <- p.Individuals[i]
	}
	close(individuals)
	wg.Wait()
}

func (p *Population) singleNextGenerationIteration(ps ParentSelector, b Crossover, m Mutation, numP int, rnd Rand) Chromosome {
	parents := make([]Chromosome, numP)
	for j := 0; j < numP; j++ {
		parents[j] = p.Individuals[ps.PickParent(p.Individuals, rnd)]
	}
	newChild := b.Reproduce(parents, rnd)
	mutatedChild := m.Mutate(newChild, rnd)
	return mutatedChild
}

func (p *Population) NextGeneration(ps ParentSelector, b Crossover, m Mutation, numP int, numGoroutine int, rnd RandForIndex) Population {
	p.calculateFitness(numGoroutine)
	ret := Population{
		Individuals: make([]Chromosome, len(p.Individuals)),
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
				ret.Individuals[idx] = p.singleNextGenerationIteration(ps, b, m, numP, rnd.Rand(idx))
			}
		}()
	}
	for i := 0; i < len(p.Individuals)-1; i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
	ret.Individuals[len(ret.Individuals)-1] = m.Mutate(p.Max(), rnd.Rand(0))
	return ret
}
