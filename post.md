# Genetic algorithm discovery of worst case Go sort inputs powered by AWS Batch and Docker

This post will give a walk thru of the following concepts:
* What are genetic algorithms
* Applying genetic algorithms to sorting inputs
* Architecting a genetic algorithm in Go
* Deploying and running a genetic algorithm at a large scale using AWS Batch and ECS

By the end of this, you should know everything you need to quickly execute
custom, large scale genetic algorithms using AWS and Go.

# What are genetic algorithms

## Better resources for genetic algorithm information
There exist really good resources online already that describe in detail how genetic algorithms work.  The two I've
found most useful are [tutorial point](https://www.tutorialspoint.com/genetic_algorithms/index.htm) and
[wikipedia](https://en.wikipedia.org/wiki/Genetic_algorithm).  My summary here is only the bare bones of genetic
algorithms since the original parts are everything else.

## Basics of genetic algorithms

You start with a solution to a problem.  This solution is called a [chromosome](https://en.wikipedia.org/wiki/Chromosome_(genetic_algorithm).

--- Put picture 

Next, you spawn a bunch of [different solutions](https://medium.com/datadriveninvestor/population-initialization-in-genetic-algorithms-ddb037da6773)
to the same problem.  Together, all of these solutions form a
[population](https://www.tutorialspoint.com/genetic_algorithms/genetic_algorithms_population.htm).

-- Put picture of lots of problems

Once you have a population of solutions to a problem, you need a [fitness](https://en.wikipedia.org/wiki/Fitness_function) function
that tells you how good a solution is.

--- Picture describing each above problem and how good or bad it was.

Now make baby solutions!  To start, find two parent solutions.  How you pick your parent solutions is called
[parent selection](https://www.tutorialspoint.com/genetic_algorithms/genetic_algorithms_parent_selection.htm). Just like natural selection, you want to bias 
to picking good solutions.  You could imagine combining the DNA of 3 or more parents, but for this example I just pick two.

--- Picture of two of the solutions.

With two parents, you need to make a child solution.  This process is called [crossover](https://en.wikipedia.org/wiki/Crossover_(genetic_algorithm).
Your child solution should be some combination of the parents.  There are [lots](https://www.tutorialspoint.com/genetic_algorithms/genetic_algorithms_crossover.htm) of ways to do this.

--- Picture of a combination solution

Finally you want to [mutate](https://en.wikipedia.org/wiki/Mutation_(genetic_algorithm) your solution.  Mutation lets you
stumble upon great solutions.  Just like for animals, mutation should be rare and may not even happen for all children.

--- Picture of mutation

Repeat this process a bunch of times until you have a new population.

--- Picture of mutated population

The number of solutions to your problem has now grown.  You need to kill off solutions to keep your population in check.
How you do this is called [survivor selection](https://en.wikipedia.org/wiki/Selection_(genetic_algorithm).  Maybe you
kill off the older solutions, maybe you make the solutions fight to the death with each other.  Really your call.

--- Picture of survived solutions

At this point you should have a population of solutions that is slightly better than your previous solutions.  It's some
combination of how you started, with a bit of mutation.  Repeat this process as much as you want.  When you decide
to stop is called your [termination condition](https://www.tutorialspoint.com/genetic_algorithms/genetic_algorithms_termination_condition.htm).

--- Picture of a bunch of strange solutions

When you stop, you find the best solution left and that's the evolved answer to your problem.  The field of genetic algorithms
and machine learning is way deeper than few paragraphs, but I hope this gives you a general sense of how it works.

## When to apply Genetic Algorithm

Genetic algorithms work well where it is *computationally prohibitive* to find the best answer to a problem.  This
includes small datasets where trying all solutions grows quickly, [like a deck of cards](https://www.mcgill.ca/oss/article/did-you-know-infographics/there-are-more-ways-arrange-deck-cards-there-are-atoms-earth),
or solutions on large datasets that have quadratic optimal solutions (even a million becomes [huge](http://www.pagetutor.com/trillion/index.html) when squared.
).

Genetic algorithms also work well when analyzing something that you're either not allowed to inspect, like a [black box](https://en.wikipedia.org/wiki/Black_box)
or problems that are [beyond](https://en.wikipedia.org/wiki/Laplace%27s_demon) our [current](https://en.wikipedia.org/wiki/Uncertainty_principle) understanding.


# Applying genetic algorithms to sorting inputs

Go's sort documentation is [short](https://golang.org/pkg/sort/#Sort) and the [code](https://github.com/golang/go/blob/go1.12.5/src/sort/sort.go#L183)
isn't too long and is worth a read.  The implementation is a combination of
* [Quicksort](https://en.wikipedia.org/wiki/Quicksort) in the normal case with [ninther](https://www.johndcook.com/blog/2009/06/23/tukey-median-ninther/) for median selection
* [Shellsort](https://en.wikipedia.org/wiki/Shellsort) when the list or segment size is small
* [Heapsort](https://en.wikipedia.org/wiki/Heapsort) if quicksort recurses too much

There exist [antiquicksort](https://www.cs.dartmouth.edu/~doug/mdmspe.pdf) algorithms to find worse case quicksort inputs,
and the go sort tests [use them](https://github.com/golang/go/blob/go1.12.5/src/sort/sort_test.go#L458).  It's not
guaranteed to produce worse case inputs for Go's case since Go uses a combination of sorting methods, but it will find
inputs that break down pretty bad.  We could reverse engineer Go's **current** sort implementation to find a bad input,
but for this problem we will use a genetic algorithm and treat Go's sort as a black box with inputs and outputs.  To start,
let's define genetic algorithm terms in the context of finding a worse case sort input.

## Chromosome 

A chromosome is a list of numbers to be sorted.  For example, `[1, 6, 3, 4, 5, 2]`.

## Fitness

The fitness of a chromosome is how many comparison operations are used in the sort.  Go's implementation guarantees `O(n*log(n))`.
In the example case, `[1, 6, 3, 4, 5, 2]` is sorted in 4 comparisons by Go, so the fitness of that array is 4.

## Parent selection

For our case, we will use K-3 [tournament selection](https://en.wikipedia.org/wiki/Tournament_selection) with p=1.  That
means we find 3 chromosomes, and the best of the 3 becomes a parent.

## Crossover

We will use [single point crossover](https://en.wikipedia.org/wiki/Crossover_(genetic_algorithm)#Single-point_crossover)
by picking a random point in each parent and spawning a child with half the array from one parent and half from another.

For example, the parents `[1, 6, 3, 4, 5, 2]` and `[6, 4, 3, 5, 2, 1]`, if crossed over at index `1` I would get
array `[1, 6, 3, 5, 2, 1]`.

## Mutate

For mutate, we'll just randomly change an index in the array.  We can do this with `1/m` ratio where `m` increases slowly
over time as we fail to improve our fitness.  For example, the array `[6, 4, 3, 5, 2, 1]` may mutate to `[6, 4, 3, 5, 10, 1]`
by changing the 2 to 10.

# Architecting a genetic algorithm in Go

[Python](https://www.python.org/) is a commonly used langauge for for machine learning and data science, especially 
combined with [NumPy](https://www.numpy.org/).  Python is perfectly fine, but I like Go's speed, static typing, and
language structure and use it for most applications.

## Code layout

You'll want to separate your core genetic algorithms from your genotype.  If you don't have too many genetic algorithm
variants, a [single package](https://github.com/cep21/geneticsort/tree/master/genetic) with separate files for each
genetic algorithm term (mutator.go, population.go, etc) is enough.  Create a
[separate package](https://github.com/cep21/geneticsort/tree/master/internal/arraysort) for each chromosome.

Your [genetic algorithm](https://github.com/cep21/geneticsort/blob/master/genetic/algorithm.go#L5) code should just
run with injections for each genetic algorithm concept.  This process is
called [dependency injection](https://en.wikipedia.org/wiki/Dependency_injection).

## Configuration

Load configuration directly with [environment variables](https://github.com/cep21/geneticsort/blob/master/main.go#L33).
This will make it easier to later rerun your same code with AWS batch, configuring your batch job with environment variables.

## Using goroutine parallelism

Genetic algorithms are very parallelizable.  When you're calculating the fitness of each individual, that can happen in
multiple [goroutines](https://tour.golang.org/concurrency/1).  An easy way to iterate over an array in parallel is to use
channels of indexes.

```go
	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	idxChan := make(chan int)
	for i := 0; i < numGoroutine; i++ {
		go func() {
			defer wg.Done()
			for idx := range idxChan {
				p.Individuals[idx].Fitness()
			}
		}()
	}
	for i := 0; i < len(p.Individuals); i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
```

Here we spawn some number of goroutines that `range` for indexes in an array to process.  We can then feed indexes to a
channel and the goroutines can process these indexes.  Once we've fed all the indexes to a channel, we `close` the channel
to tell the `for .. range` inside the goroutine to stop.  Finally, we `wg.Wait` to block until all our goroutines are done.

We can select children for the next generation in a similarly parallel way.  This works because we don't mutate the working
set while picking parents and making children for the next generation.

```go
	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	idxChan := make(chan int)
	for i := 0; i < numGoroutine; i++ {
		go func() {
			defer wg.Done()
			for idx := range idxChan {
				ret.Individuals[idx] = p.singleNextGenerationIteration(ps, b, m, numP, rnd.Rand(idx))
			}
		}()
	}
	for i := 0; i < len(p.Individuals)-1; i++ {
		idxChan <- i
	}
	close(idxChan)
	wg.Wait()
```

## Rand in parallel algorithms

## Creating a Docker container of your Go program

# Deploying and running a genetic algorithm at a large scale using AWS Batch and ECS

The last 