package genetic

import "fmt"

type SurvivorSelection interface {
	NextGeneration(previous *Population, candidate *Population, r Rand) Population
	String() string
}

type ParentSurvivorSelection struct {
	ParentSelector ParentSelector
}

func (p *ParentSurvivorSelection) String() string {
	return fmt.Sprintf("parent-select-%s", p.ParentSelector.String())
}

func (p *ParentSurvivorSelection) NextGeneration(previous *Population, candidate *Population, r Rand) Population {
	allParents := make([]Chromosome, 0, len(previous.Individuals)+len(candidate.Individuals))
	allParents = append(allParents, previous.Individuals...)
	allParents = append(allParents, candidate.Individuals...)
	var ret Population
	alreadyPickedParents := make(map[int]struct{})
	for i := 0; i < (len(previous.Individuals)+len(candidate.Individuals))/2; i++ {
		newIdx := p.ParentSelector.PickParent(allParents, r)
		// We don't want the same person twice in the next generation
		// We don't want to update allParents each iteration, since that turns this from O(N) to O(N^2)
		// So we kinda cheat and only update allParents if we pick the same index twice
		if _, contains := alreadyPickedParents[newIdx]; contains {
			// Remove all alreadyPickedParents from allParents and reset alreadyPickedParents
			newAllParents := make([]Chromosome, 0, len(allParents))
			for i := range allParents {
				if _, contains := alreadyPickedParents[i]; !contains {
					newAllParents = append(newAllParents, allParents[i])
				}
			}
			allParents = newAllParents
			alreadyPickedParents = make(map[int]struct{})
			newIdx = p.ParentSelector.PickParent(allParents, r)
		}
		ret.Individuals = append(ret.Individuals, allParents[newIdx])
	}
	return ret
}

var _ SurvivorSelection = &ParentSurvivorSelection{}
