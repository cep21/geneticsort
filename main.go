package main

import (
	"fmt"
	"github.com/cep21/geneticsort/genetic"
	"github.com/cep21/geneticsort/internal/arraysort"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	r := &genetic.LockedRand{G: rand.New(rand.NewSource(0))}
	a := genetic.Algorithm{
		ParentSelector: &genetic.TournamentParentSelector {
			R: r,
			//K: 4,
		},
		Factory: &arraysort.ArraySortingFactory {
			R: r,
			// 100 is 1332
			// 500 is 13989
			IndividualSize: 100,
		},
		Terminator: &genetic.TimingExecutor{
			Duration: time.Minute,
		},
		Breeder: &genetic.SplitReproduce{
			R: r,
		},
		Mutator: &genetic.LookAheadMutator{
			R: r,
			MutationRatio: 10,
		},
		NumberOfParents: 2,
		PopulationSize: 5000,
		NumGoroutine: runtime.NumCPU(),
	}
	fittest := a.Run()
	fittest.(genetic.Simplifyable).Simplify()
	fmt.Println(fittest)
}
