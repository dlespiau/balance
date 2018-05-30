package main

import (
	"math"
	"math/rand"
)

// Rander are objects that can produce random numbers.
type Rander interface {
	Rand() float64
}

// Constant generates a constant number.
type Constant struct {
	Number float64
}

// Rand implements Rander.
func (c *Constant) Rand() float64 {
	return c.Number
}

// LogNormal generates random numbers that are Log Normal distributed.
// https://en.wikipedia.org/wiki/Log-normal_distribution
type LogNormal struct {
	Mu, Sigma float64
}

// Rand implements Rander.
func (l *LogNormal) Rand() float64 {
	r := rand.NormFloat64()
	return math.Exp(r*l.Sigma + l.Mu)
}
