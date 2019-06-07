package genetic

import (
	"fmt"
	"strings"
	"time"
)

type ExecutionTerminator interface {
	StopExecution(p Population, r Rand) bool
	String() string
}

type CountingExecutor struct {
	Limit int
	i     int
}

func (c *CountingExecutor) String() string {
	return fmt.Sprintf("counting-%d", c.Limit)
}

func (c *CountingExecutor) StopExecution(p Population, _ Rand) bool {
	if c.i >= c.Limit {
		return true
	}
	c.i++
	return false
}

type MultiStopExecutor struct {
	Executors []ExecutionTerminator
}

func (c *MultiStopExecutor) String() string {
	parts := make([]string, 0, len(c.Executors))
	for _, e := range c.Executors {
		parts = append(parts, e.String())
	}
	return fmt.Sprintf("multi-%s", strings.Join(parts, ","))
}

func (c *MultiStopExecutor) StopExecution(p Population, r Rand) bool {
	shouldStop := false
	for _, e := range c.Executors {
		shouldStop = shouldStop || e.StopExecution(p, r)
	}
	return shouldStop
}

var _ ExecutionTerminator = &MultiStopExecutor{}

type NoImprovementExecutor struct {
	Consecutive        int
	currentBest        int
	currentConsecutive int
}

func (c *NoImprovementExecutor) String() string {
	return fmt.Sprintf("consecutive-%d", c.Consecutive)
}

func (c *NoImprovementExecutor) StopExecution(p Population, _ Rand) bool {
	best := p.Max().Fitness()
	if best > c.currentBest {
		c.currentBest = best
		c.currentConsecutive = 0
	} else {
		c.currentConsecutive++
		if c.currentConsecutive > c.Consecutive {
			return true
		}
	}
	return false
}

var _ ExecutionTerminator = &NoImprovementExecutor{}

var _ ExecutionTerminator = &CountingExecutor{}

type TimingExecutor struct {
	Duration  time.Duration
	startTime time.Time
	Now       func() time.Time
}

func (c *TimingExecutor) String() string {
	return fmt.Sprintf("timing-%s", c.Duration.String())
}

func (c *TimingExecutor) now() time.Time {
	if c.Now == nil {
		return time.Now()
	}
	return c.Now()
}

func (c *TimingExecutor) StopExecution(p Population, _ Rand) bool {
	if c.startTime.IsZero() {
		c.startTime = c.now()
	}
	curTime := c.now()
	return curTime.Sub(c.startTime) > c.Duration
}

var _ ExecutionTerminator = &TimingExecutor{}
