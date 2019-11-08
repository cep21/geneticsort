package genetic

import (
	"fmt"
	"strings"
	"time"
)

type Termination interface {
	StopExecution(p Population, r Rand) bool
	String() string
}

type CountingTermination struct {
	Limit int
	i     int
}

func (c *CountingTermination) String() string {
	return fmt.Sprintf("counting-%d", c.Limit)
}

func (c *CountingTermination) StopExecution(p Population, _ Rand) bool {
	if c.i >= c.Limit {
		return true
	}
	c.i++
	return false
}

type MultiTermination struct {
	Executors []Termination
}

func (c *MultiTermination) String() string {
	parts := make([]string, 0, len(c.Executors))
	for _, e := range c.Executors {
		parts = append(parts, e.String())
	}
	return fmt.Sprintf("multi-%s", strings.Join(parts, ","))
}

func (c *MultiTermination) StopExecution(p Population, r Rand) bool {
	shouldStop := false
	for _, e := range c.Executors {
		shouldStop = shouldStop || e.StopExecution(p, r)
	}
	return shouldStop
}

var _ Termination = &MultiTermination{}

type NoImprovementTermination struct {
	Consecutive        int
	currentBest        int
	currentConsecutive int
}

func (c *NoImprovementTermination) String() string {
	return fmt.Sprintf("consecutive-%d", c.Consecutive)
}

func (c *NoImprovementTermination) StopExecution(p Population, _ Rand) bool {
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

var _ Termination = &NoImprovementTermination{}

var _ Termination = &CountingTermination{}

type TimingTermination struct {
	Duration  time.Duration
	startTime time.Time
	Now       func() time.Time
}

func (c *TimingTermination) String() string {
	return fmt.Sprintf("timing-%s", c.Duration.String())
}

func (c *TimingTermination) now() time.Time {
	if c.Now == nil {
		return time.Now()
	}
	return c.Now()
}

func (c *TimingTermination) StopExecution(p Population, _ Rand) bool {
	if c.startTime.IsZero() {
		c.startTime = c.now()
	}
	curTime := c.now()
	return curTime.Sub(c.startTime) > c.Duration
}

var _ Termination = &TimingTermination{}
