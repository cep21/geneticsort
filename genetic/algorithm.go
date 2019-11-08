package genetic

import (
	"log"
)

type Algorithm struct {
	Log               *log.Logger
	RandForIndex      RandForIndex
	ParentSelector    ParentSelector
	Factory           ChromosomeFactory
	Terminator        Termination
	Crossover         Crossover
	SurvivorSelection SurvivorSelection
	Mutator           Mutation
	NumberOfParents   int
	PopulationSize    int
	NumGoroutine      int
}

func (a *Algorithm) Run() Chromosome {
	currentPopulation := SpawnPopulation(a.PopulationSize, a.Factory, a.RandForIndex.Rand(0))
	best := currentPopulation.Max()
	asDynamic, isDynamic := a.Mutator.(DynamicMutation)
	if isDynamic {
		asDynamic.ResetMutationRate(a.RandForIndex.Rand(0))
	}
	runCounter := 0
	for {
		if a.Log != nil {
			a.Log.Println("Index/mean/max", runCounter, currentPopulation.Average(), currentPopulation.Max().Fitness())
		}
		runCounter++
		if a.Terminator.StopExecution(currentPopulation, a.RandForIndex.Rand(0)) {
			if asSimpl, canSimpl := best.(Simplifyable); canSimpl {
				asSimpl.Simplify()
			}
			return best
		}
		nextPopulation := currentPopulation.NextGeneration(a.ParentSelector, a.Crossover, a.Mutator, a.NumberOfParents, a.NumGoroutine, a.RandForIndex)
		nextPopulation = a.SurvivorSelection.NextGeneration(&currentPopulation, &nextPopulation, a.RandForIndex.Rand(0))
		nextBest := nextPopulation.Max()
		if best.Fitness() < nextBest.Fitness() {
			best = nextPopulation.Max()
			if isDynamic {
				asDynamic.ResetMutationRate(a.RandForIndex.Rand(0))
			}
		} else if isDynamic {
			asDynamic.IncreaseMutationRate(a.RandForIndex.Rand(0))
		}
		currentPopulation = nextPopulation
	}
}
