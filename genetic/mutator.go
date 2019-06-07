package genetic

import "fmt"

type Mutator interface {
	Mutate(in Individual, r Rand) Individual
	String() string
}

type SwapMutator struct {
}

func (a *SwapMutator) String() string {
	return fmt.Sprintf("swap")
}

func (a *SwapMutator) Mutate(in Individual, r Rand) Individual {
	asSwap, ok := in.(Array)
	if !ok {
		panic("Trying to do swap mutation with something that isn't swappable")
	}

	i := r.Intn(asSwap.Len())
	j := r.Intn(asSwap.Len())
	if i == j {
		return in
	}
	asS := asSwap.Clone()
	asS.(Array).Swap(i, j)
	return asS
}

type DynamicMutation interface {
	Mutator
	IncreaseMutationRate(r Rand)
	ResetMutationRate(r Rand)
}

var _ Mutator = &SwapMutator{}

type LookAheadMutator struct {
}

func (a *LookAheadMutator) String() string {
	return "lookahead"
}

func (a *LookAheadMutator) Mutate(in Individual, r Rand) Individual {
	asSwap, ok := in.(Array)
	if !ok {
		panic("Trying to do swap mutation with something that isn't swappable")
	}
	newPerson := asSwap.Clone().(Array)
	startingIndex := r.Intn(asSwap.Len())
	for i := 0; i < asSwap.Len(); i++ {
		shouldSwap := r.Intn(asSwap.Len()) == 0
		if !shouldSwap {
			continue
		}
		swapWith := r.Intn(asSwap.Len())
		if swapWith == i {
			continue
		}
		newPerson.Swap((i+startingIndex)%asSwap.Len(), (swapWith+startingIndex)%asSwap.Len())
	}
	return newPerson
}

type PassThruDynamicMutation struct {
	PassTo               Mutator
	MutationRatio        int
	currentMutationRatio int
}

func (p *PassThruDynamicMutation) Mutate(in Individual, r Rand) Individual {
	if r.Intn(p.currentMutationRatio) != 0 {
		return in
	}
	return p.PassTo.Mutate(in, r)
}

func (p *PassThruDynamicMutation) String() string {
	return fmt.Sprintf("pass-%d-%s", p.MutationRatio, p.PassTo.String())
}

func (p *PassThruDynamicMutation) IncreaseMutationRate(r Rand) {
	p.currentMutationRatio--
	if p.currentMutationRatio <= 1 {
		p.currentMutationRatio = 1
	}
}

func (p *PassThruDynamicMutation) ResetMutationRate(r Rand) {
	p.currentMutationRatio = p.MutationRatio
}

type IndexMutation struct {
}

func (i *IndexMutation) Mutate(in Individual, r Rand) Individual {
	ret := in.Clone()
	asArray := ret.(Array)
	asArray.Randomize(r.Intn(asArray.Len()), r)
	return ret
}

func (i *IndexMutation) String() string {
	return "index-mutation"
}

var _ Mutator = &LookAheadMutator{}
var _ Mutator = &LookAheadMutator{}
var _ Mutator = &IndexMutation{}
var _ DynamicMutation = &PassThruDynamicMutation{}
