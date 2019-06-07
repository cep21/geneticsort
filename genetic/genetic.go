package genetic

type Individual interface {
	Fitness() int
	Clone() Individual
	Shell() Individual
	String() string
}

type LocalOptimization interface {
	LocallyOptimize() Individual
}

type Simplifyable interface {
	Simplify()
}

type Array interface {
	Individual
	Swap(i, j int)
	Copy(from Array, start int, end int, into int)
	Randomize(int, Rand)
	Len() int
}

type IndividualFactory interface {
	Spawn(G Rand) Individual
	Family() string
}
