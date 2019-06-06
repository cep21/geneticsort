package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/cep21/geneticsort/genetic"
	"github.com/cep21/geneticsort/internal/arraysort"
)

type runConfig struct {
	ArraySize      int
	KTournament    int
	Duration       time.Duration
	MutationRation int
	PopulationSize int
}

func load() runConfig {
	var ret runConfig
	ret.ArraySize = mustOsInt("ARRAY_SIZE", 1000)
	ret.KTournament = mustOsInt("K_TOURNAMENT", 3)
	ret.MutationRation = mustOsInt("MUTATION_RATION", 30)
	ret.PopulationSize = mustOsInt("POPULATION_SIZE", 1000)
	ret.Duration = mustOsDur("RUN_TIME", time.Minute)
	return ret
}

// must_ie return i or panics if err.  the "ie" stands for "int/error"
func mustOsInt(s string, defaultVal int) int {
	a := os.Getenv(s)
	if a == "" {
		return defaultVal
	}
	ret, err := strconv.Atoi(a)
	if err != nil {
		panic(err)
	}
	return ret
}

// must_ie return i or panics if err.  the "ie" stands for "int/error"
func mustOsDur(s string, defaultVal time.Duration) time.Duration {
	a := os.Getenv(s)
	if a == "" {
		return defaultVal
	}
	ret, err := time.ParseDuration(a)
	if err != nil {
		panic(err)
	}
	return ret
}

func main() {
	r := &genetic.LockedRand{G: rand.New(rand.NewSource(0))}
	conf := load()
	a := genetic.Algorithm{
		Log: log.New(os.Stdout, "", log.LstdFlags),
		ParentSelector: &genetic.TournamentParentSelector{
			R: r,
			K: conf.KTournament,
		},
		Factory: &arraysort.ArraySortingFactory{
			R: r,
			// According to go stdlib TestAdversary
			// - 100 is 1332
			// - 500 is 13989
			// - 1000 is 33454
			IndividualSize: conf.ArraySize,
		},
		Terminator: &genetic.TimingExecutor{
			Duration: conf.Duration,
		},
		Breeder: &genetic.SplitReproduce{
			R: r,
		},
		Mutator: &genetic.LookAheadMutator{
			R:             r,
			MutationRatio: conf.MutationRation,
		},
		NumberOfParents: 2,
		PopulationSize:  conf.PopulationSize,
		NumGoroutine:    runtime.NumCPU(),
	}
	fittest := a.Run()
	fittest.(genetic.Simplifyable).Simplify()
	fmt.Println(fittest)
}
