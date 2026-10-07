package math

import "math"

// Complex is a minimal complex number (avoids importing math/cmplx everywhere).
type Complex struct{ Re, Im float64 }

func (c Complex) Abs() float64 { return math.Hypot(c.Re, c.Im) }

// FFT computes the in-place radix-2 Cooley–Tukey transform. The input length
// must be a power of two — use Pad first if it isn't. inverse=false is the
// forward transform.
func FFT(x []Complex, inverse bool) {
	n := len(x)
	if n == 0 || n&(n-1) != 0 {
		panic("bones: FFT length must be a power of two")
	}
	// Bit-reversal permutation.
	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j &= ^bit
		}
		j |= bit
		if i < j {
			x[i], x[j] = x[j], x[i]
		}
	}
	// Butterflies.
	for length := 2; length <= n; length <<= 1 {
		ang := 2 * math.Pi / float64(length)
		if !inverse {
			ang = -ang
		}
		wlen := Complex{math.Cos(ang), math.Sin(ang)}
		for i := 0; i < n; i += length {
			w := Complex{1, 0}
			for k := 0; k < length/2; k++ {
				u := x[i+k]
				v := cmul(x[i+k+length/2], w)
				x[i+k] = Complex{u.Re + v.Re, u.Im + v.Im}
				x[i+k+length/2] = Complex{u.Re - v.Re, u.Im - v.Im}
				w = cmul(w, wlen)
			}
		}
	}
	if inverse {
		for i := range x {
			x[i].Re /= float64(n)
			x[i].Im /= float64(n)
		}
	}
}

// Spectrum returns the magnitude spectrum of a real signal, zero-padding up to
// the next power of two. The result has len = paddedN/2 (the non-redundant half
// for a real input).
func Spectrum(signal Vec) Vec {
	n := nextPow2(len(signal))
	if n < 2 {
		return Vec{}
	}
	buf := make([]Complex, n)
	for i, v := range signal {
		buf[i] = Complex{v, 0}
	}
	FFT(buf, false)
	out := make(Vec, n/2)
	for i := 0; i < n/2; i++ {
		out[i] = buf[i].Abs()
	}
	return out
}

// Pad returns a copy of signal zero-extended to the next power of two.
func Pad(signal Vec) Vec {
	n := nextPow2(len(signal))
	out := make(Vec, n)
	copy(out, signal)
	return out
}

func cmul(a, b Complex) Complex {
	return Complex{a.Re*b.Re - a.Im*b.Im, a.Re*b.Im + a.Im*b.Re}
}

func nextPow2(n int) int {
	if n <= 1 {
		return n
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}