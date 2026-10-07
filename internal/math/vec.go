// Package math is the glyphai's skeleton: the exact numerical math the rest of the
// body stands on — vectors, parallel matrix multiply, FFT, wavelets and the
// neural-net activations. No external dependencies; everything is float64 and
// deterministic so spells and the cerebellum agree to the bit.
package math

import "math"

// Vec is a dense float64 vector.
type Vec []float64

// Dot is the inner product. Panics if lengths differ.
func Dot(a, b Vec) float64 {
	if len(a) != len(b) {
		panic("bones: Dot length mismatch")
	}
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

// Norm is the Euclidean length.
func Norm(a Vec) float64 { return math.Sqrt(Dot(a, a)) }

// Cosine similarity in [-1,1]; 0 if either vector is empty.
func Cosine(a, b Vec) float64 {
	na, nb := Norm(a), Norm(b)
	if na == 0 || nb == 0 {
		return 0
	}
	return Dot(a, b) / (na * nb)
}

// Add returns a+b.
func Add(a, b Vec) Vec {
	out := make(Vec, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

// Scale returns a*s.
func Scale(a Vec, s float64) Vec {
	out := make(Vec, len(a))
	for i := range a {
		out[i] = a[i] * s
	}
	return out
}

// AXPY computes y += a*x in place (the classic fused op).
func AXPY(a float64, x, y Vec) {
	for i := range x {
		y[i] += a * x[i]
	}
}

// Mean of a vector (0 for empty).
func Mean(a Vec) float64 {
	if len(a) == 0 {
		return 0
	}
	var s float64
	for _, v := range a {
		s += v
	}
	return s / float64(len(a))
}

// ArgMax returns the index of the largest element (-1 if empty).
func ArgMax(a Vec) int {
	if len(a) == 0 {
		return -1
	}
	bi, bv := 0, a[0]
	for i, v := range a {
		if v > bv {
			bi, bv = i, v
		}
	}
	return bi
}