package math

import "math"

// Wavelets give glyphai multi-scale perception: they split a signal into coarse
// approximation + fine detail at successive scales, which is exactly how you
// see structure in an image or audio "at a glance and up close" at once.

// HaarForward runs one level of the Haar discrete wavelet transform. It returns
// the approximation (averages) and detail (differences) coefficients, each of
// length len(x)/2. len(x) must be even.
func HaarForward(x Vec) (approx, detail Vec) {
	n := len(x) / 2
	approx = make(Vec, n)
	detail = make(Vec, n)
	const s = math.Sqrt2
	for i := 0; i < n; i++ {
		a := x[2*i]
		b := x[2*i+1]
		approx[i] = (a + b) / s
		detail[i] = (a - b) / s
	}
	return
}

// HaarInverse reconstructs the signal from one level of Haar coefficients.
func HaarInverse(approx, detail Vec) Vec {
	n := len(approx)
	out := make(Vec, 2*n)
	const s = math.Sqrt2
	for i := 0; i < n; i++ {
		a := approx[i]
		d := detail[i]
		out[2*i] = (a + d) / s
		out[2*i+1] = (a - d) / s
	}
	return out
}

// HaarDecompose runs the full multi-level Haar transform, returning the final
// approximation followed by detail bands from finest to coarsest. levels is
// clamped so each level has an even length.
func HaarDecompose(x Vec, levels int) (coeffs []Vec) {
	cur := append(Vec(nil), x...)
	for l := 0; l < levels && len(cur) >= 2 && len(cur)%2 == 0; l++ {
		a, d := HaarForward(cur)
		coeffs = append(coeffs, d)
		cur = a
	}
	coeffs = append(coeffs, cur) // final approximation last
	return
}

// Daubechies-4 (db2) coefficients: a smoother basis than Haar, better for
// natural signals. One forward level with periodic boundary handling.
// math.Sqrt3 isn't in the std lib, so the root of 3 is a literal constant.
const sqrt3 = 1.7320508075688772

var (
	db4h = [4]float64{
		(1 + sqrt3) / (4 * math.Sqrt2),
		(3 + sqrt3) / (4 * math.Sqrt2),
		(3 - sqrt3) / (4 * math.Sqrt2),
		(1 - sqrt3) / (4 * math.Sqrt2),
	}
	db4g = [4]float64{db4h[3], -db4h[2], db4h[1], -db4h[0]}
)

// Daub4Forward runs one level of the Daubechies-4 transform with periodic
// extension. len(x) must be even and >= 4.
func Daub4Forward(x Vec) (approx, detail Vec) {
	n := len(x)
	half := n / 2
	approx = make(Vec, half)
	detail = make(Vec, half)
	for i := 0; i < half; i++ {
		var a, d float64
		for k := 0; k < 4; k++ {
			idx := (2*i + k) % n
			a += db4h[k] * x[idx]
			d += db4g[k] * x[idx]
		}
		approx[i] = a
		detail[i] = d
	}
	return
}