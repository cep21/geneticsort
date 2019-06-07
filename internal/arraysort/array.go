package arraysort

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cep21/geneticsort/genetic"
)

type arraySortingIndividual struct {
	vals    []int
	fitness *int
}

func (c *arraySortingIndividual) Randomize(idx int, r genetic.Rand) {
	c.vals[idx] = r.Int()
}

func (c *arraySortingIndividual) Copy(from genetic.Array, start int, end int, into int) {
	asS := from.(*arraySortingIndividual)
	copy(c.vals[into:], asS.vals[start:end])
}

func (c *arraySortingIndividual) Simplify() {
	tmpVals := make([]int, len(c.vals))
	copy(tmpVals, c.vals)
	comparisons := 0
	sort.Slice(tmpVals, func(i, j int) bool {
		comparisons++
		return tmpVals[i] < tmpVals[j]
	})
	valToIdx := make(map[int]int)
	for idx, v := range tmpVals {
		valToIdx[v] = idx
	}
	for idx, v := range c.vals {
		c.vals[idx] = valToIdx[v]
	}
}

func (c *arraySortingIndividual) Len() int {
	return len(c.vals)
}

func mustPrint(_ int, err error) {
	if err != nil {
		panic(err)
	}
}

var _ genetic.Individual = &arraySortingIndividual{}
var _ genetic.Array = &arraySortingIndividual{}

func (c *arraySortingIndividual) String() string {
	var s strings.Builder
	for i := 0; i < len(c.vals); i++ {
		if i != 0 {
			mustPrint(s.WriteString(","))
		}
		mustPrint(fmt.Fprintf(&s, "%d", c.vals[i]))
	}
	return s.String()
}

type ArraySortingFactory struct {
	IndividualSize int
}

var _ genetic.IndividualFactory = &ArraySortingFactory{}

func (a *ArraySortingFactory) Family() string {
	return fmt.Sprintf("intarray-sort-%d", a.IndividualSize)
}

func (a *ArraySortingFactory) Spawn(r genetic.Rand) genetic.Individual {
	c := &arraySortingIndividual{
		vals: make([]int, a.IndividualSize),
	}
	for i := 0; i < a.IndividualSize; i++ {
		c.vals[i] = r.Int()
	}
	return c
}

func (c *arraySortingIndividual) Shell() genetic.Individual {
	return &arraySortingIndividual{
		vals: make([]int, len(c.vals)),
	}
}

func (c *arraySortingIndividual) Clone() genetic.Individual {
	ret := &arraySortingIndividual{
		vals: make([]int, len(c.vals)),
	}
	copy(ret.vals, c.vals)
	return ret
}

func (c *arraySortingIndividual) Fitness() int {
	if c.fitness != nil {
		return *c.fitness
	}

	tmpVals := make([]int, len(c.vals))
	copy(tmpVals, c.vals)
	comparisons := 0
	sort.Slice(tmpVals, func(i, j int) bool {
		comparisons++
		return tmpVals[i] < tmpVals[j]
	})
	c.fitness = &comparisons
	return comparisons
}

func (c *arraySortingIndividual) MustBeSorted() {
	if !sort.IntsAreSorted(c.vals[0:]) {
		panic("Nope not sorted")
	}
}

func (c *arraySortingIndividual) Swap(i, j int) {
	c.vals[i], c.vals[j] = c.vals[j], c.vals[i]
}
