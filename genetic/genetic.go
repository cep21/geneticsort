package genetic

type Chromosome interface {
	Fitness() int
	Clone() Chromosome
	Shell() Chromosome
	String() string
}

type LocalOptimization interface {
	LocallyOptimize() Chromosome
}

type Simplifyable interface {
	Simplify()
}

type Array interface {
	Chromosome
	Swap(i, j int)
	Copy(from Array, start int, end int, into int)
	Randomize(int, Rand)
	Len() int
}

type ChromosomeFactory interface {
	Spawn(G Rand) Chromosome
	Family() string
}
