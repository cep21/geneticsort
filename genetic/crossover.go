package genetic

import "fmt"

type Crossover interface {
	Reproduce(in []Chromosome, r Rand) Chromosome
	String() string
}

type OnePointCrossover struct {
}

func (s *OnePointCrossover) String() string {
	return fmt.Sprintf("split")
}

func (s *OnePointCrossover) Reproduce(in []Chromosome, r Rand) Chromosome {
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
	midPoint := r.Intn(asSwap.Len())
	prefC := r.Intn(2)%2 == 0
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

var _ Crossover = &OnePointCrossover{}
