package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cep21/geneticsort/internal/record"
	"github.com/cep21/geneticsort/internal/record/dynamorecord"

	"github.com/cep21/geneticsort/genetic"
	"github.com/cep21/geneticsort/internal/arraysort"
)

type runConfig struct {
	ArraySize        int
	KTournament      int
	Duration         time.Duration
	MutationRation   int
	PopulationSize   int
	Seed             int64
	TerminationStall int
	DynamoDBTable    string
}

func load() runConfig {
	var ret runConfig
	ret.ArraySize = mustOsInt("ARRAY_SIZE", 1000)
	ret.KTournament = mustOsInt("K_TOURNAMENT", 3)
	ret.MutationRation = mustOsInt("MUTATION_RATION", 30)
	ret.PopulationSize = mustOsInt("POPULATION_SIZE", 1000)
	ret.TerminationStall = mustOsInt("TERMINATE_ON_STALL", 50)
	ret.Seed = mustOsInt64("RAND_SEED", 0)
	if ret.Seed < 0 {
		ret.Seed = time.Now().UnixNano()
	}
	ret.Duration = mustOsDur("RUN_TIME", time.Minute)
	ret.DynamoDBTable = os.Getenv("DYNAMODB_TABLE")
	return ret
}

func mustOsInt(s string, defaultVal int) int {
	return int(mustOsInt64(s, int64(defaultVal)))
}

// must_ie return i or panics if err.  the "ie" stands for "int/error"
func mustOsInt64(s string, defaultVal int64) int64 {
	a := os.Getenv(s)
	if a == "" {
		return defaultVal
	}
	ret, err := strconv.ParseInt(a, 10, 64)
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
	conf := load()
	a := genetic.Algorithm{
		RandForIndex: genetic.ArrayRandForIdx(conf.PopulationSize, conf.Seed, func(seed int64) genetic.Rand {
			return rand.New(rand.NewSource(seed))
		}),
		Log: log.New(os.Stdout, "", log.LstdFlags),
		ParentSelector: &genetic.TournamentParentSelector{
			K: conf.KTournament,
		},
		Factory: &arraysort.ArraySortingFactory{
			// According to go stdlib TestAdversary
			// - 100 is 1332
			// - 500 is 13989
			// - 1000 is 33454
			IndividualSize: conf.ArraySize,
		},
		Terminator: &genetic.MultiTermination{
			Executors: []genetic.Termination{
				&genetic.TimingTermination{
					Duration: conf.Duration,
				},
				&genetic.NoImprovementTermination{
					Consecutive: conf.TerminationStall,
				},
			},
		},
		Crossover: &genetic.OnePointCrossover{},
		Mutator: &genetic.PassThruDynamicMutation{
			MutationRatio: conf.MutationRation,
			PassTo:        &genetic.IndexMutation{},
		},
		NumberOfParents: 2,
		PopulationSize:  conf.PopulationSize,
		NumGoroutine:    runtime.NumCPU(),
	}
	fittest := a.Run()
	fittest.(genetic.Simplifyable).Simplify()
	fmt.Println(fittest)
	if conf.DynamoDBTable != "" {
		ses := session.Must(session.NewSession())
		ddb := dynamodb.New(ses)
		drec := &dynamorecord.Recorder{
			Client:    ddb,
			TableName: conf.DynamoDBTable,
		}
		if err := drec.Record(context.Background(), record.Record{
			Algorithm:     a,
			BestCandidate: fittest,
		}); err != nil {
			panic(err)
		}
	}
}
