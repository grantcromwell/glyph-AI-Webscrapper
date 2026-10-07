// Package thalamus is Argos's sensory-statistics lobe. It fits Gaussians over the
// dog's OWN glyph/logogram measurements — distances between memories, richness of
// senses — so that judgements (what is typical, what is surprising, what should
// merge) fall out of the data's distribution, not out of chosen thresholds. The
// only thing fixed here is the mathematics of the normal curve; the numbers come
// from the dog. Pure stdlib.
package routing

import "math"

// Stats is a fitted one-dimensional Gaussian.
type Stats struct {
	Mean, Std float64
	N         int
}

// Fit estimates the Gaussian (mean, std) of a sample.
func Fit(xs []float64) Stats {
	if len(xs) == 0 {
		return Stats{}
	}
	var m float64
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	var v float64
	for _, x := range xs {
		v += (x - m) * (x - m)
	}
	return Stats{Mean: m, Std: math.Sqrt(v / float64(len(xs))), N: len(xs)}
}

// Z is the standard score of x (how many σ from the mean).
func (s Stats) Z(x float64) float64 {
	if s.Std == 0 {
		return 0
	}
	return (x - s.Mean) / s.Std
}

// Surprise is the negative log-density of x under the Gaussian — how unexpected a
// reading is. Larger = more surprising. (Useful for novelty / outlier senses.)
func (s Stats) Surprise(x float64) float64 {
	if s.Std == 0 {
		return 0
	}
	z := s.Z(x)
	return 0.5*z*z + math.Log(s.Std*math.Sqrt(2*math.Pi))
}

// LowerTail is the probability a draw lands at or below x — Φ(z). For a distance,
// LowerTail(d) is the chance a pair is at least this close: small distances sit in
// the low tail and return high probability.
func (s Stats) LowerTail(x float64) float64 { return NormCDF(s.Z(x)) }

// NormCDF is the standard normal cumulative distribution Φ(z).
func NormCDF(z float64) float64 { return 0.5 * (1 + math.Erf(z/math.Sqrt2)) }