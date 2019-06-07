package genetic

import (
	"fmt"
	"math"
)

type ParentSelector interface {
	PickParent([]Individual, Rand) int
	String() string
}

type TournamentParentSelector struct {
	K int
}

func (s TournamentParentSelector) String() string {
	return fmt.Sprintf("K-Tournament-%d", s.K)
}

func (s TournamentParentSelector) PickParent(c []Individual, r Rand) int {
	k := s.K
	if k == 0 {
		k = int(math.Log(float64(len(c))) + 1)
	}
	current := r.Intn(len(c))
	for i := 1; i < k; i++ {
		other := r.Intn(len(c))
		if c[current].Fitness() < c[other].Fitness() {
			current = other
		}
	}
	return current
}

var _ ParentSelector = &TournamentParentSelector{}
