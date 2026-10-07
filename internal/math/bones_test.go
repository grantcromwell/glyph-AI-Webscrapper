package math

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestMatMulIdentity(t *testing.T) {
	a := &Mat{Rows: 2, Cols: 3, Data: Vec{1, 2, 3, 4, 5, 6}}
	id := &Mat{Rows: 3, Cols: 3, Data: Vec{1, 0, 0, 0, 1, 0, 0, 0, 1}}
	c := MatMul(a, id)
	for i := range a.Data {
		if !approx(c.Data[i], a.Data[i], 1e-9) {
			t.Fatalf("A*I != A at %d: %v vs %v", i, c.Data[i], a.Data[i])
		}
	}
}

func TestMatMulKnown(t *testing.T) {
	a := &Mat{Rows: 2, Cols: 2, Data: Vec{1, 2, 3, 4}}
	b := &Mat{Rows: 2, Cols: 2, Data: Vec{5, 6, 7, 8}}
	c := MatMul(a, b)
	want := Vec{19, 22, 43, 50}
	for i := range want {
		if !approx(c.Data[i], want[i], 1e-9) {
			t.Fatalf("MatMul wrong at %d: %v want %v", i, c.Data[i], want[i])
		}
	}
}

func TestFFTRoundTrip(t *testing.T) {
	x := []Complex{{1, 0}, {2, 0}, {3, 0}, {4, 0}, {4, 0}, {3, 0}, {2, 0}, {1, 0}}
	orig := append([]Complex(nil), x...)
	FFT(x, false)
	FFT(x, true)
	for i := range x {
		if !approx(x[i].Re, orig[i].Re, 1e-9) || !approx(x[i].Im, orig[i].Im, 1e-9) {
			t.Fatalf("FFT round trip failed at %d: %+v vs %+v", i, x[i], orig[i])
		}
	}
}

func TestFFTConstantSignal(t *testing.T) {
	// A constant signal has all energy in bin 0.
	x := []Complex{{1, 0}, {1, 0}, {1, 0}, {1, 0}}
	FFT(x, false)
	if !approx(x[0].Re, 4, 1e-9) {
		t.Fatalf("DC bin = %v, want 4", x[0].Re)
	}
	for i := 1; i < len(x); i++ {
		if !approx(x[i].Abs(), 0, 1e-9) {
			t.Fatalf("bin %d = %v, want 0", i, x[i].Abs())
		}
	}
}

func TestHaarPerfectReconstruction(t *testing.T) {
	x := Vec{4, 6, 10, 12, 3, 7, 9, 1}
	a, d := HaarForward(x)
	r := HaarInverse(a, d)
	for i := range x {
		if !approx(r[i], x[i], 1e-9) {
			t.Fatalf("Haar reconstruction failed at %d: %v vs %v", i, r[i], x[i])
		}
	}
}

func TestHaarEnergyPreserved(t *testing.T) {
	// Orthonormal transform preserves L2 energy.
	x := Vec{4, 6, 10, 12, 3, 7, 9, 1}
	a, d := HaarForward(x)
	var ex, ec float64
	for _, v := range x {
		ex += v * v
	}
	for i := range a {
		ec += a[i]*a[i] + d[i]*d[i]
	}
	if !approx(ex, ec, 1e-6) {
		t.Fatalf("Haar energy not preserved: %v vs %v", ex, ec)
	}
}

func TestDaub4SumRules(t *testing.T) {
	// Low-pass coeffs sum to sqrt(2); high-pass coeffs sum to 0.
	var sh, sg float64
	for i := 0; i < 4; i++ {
		sh += db4h[i]
		sg += db4g[i]
	}
	if !approx(sh, math.Sqrt2, 1e-9) {
		t.Fatalf("db4 low-pass sum = %v, want sqrt2", sh)
	}
	if !approx(sg, 0, 1e-9) {
		t.Fatalf("db4 high-pass sum = %v, want 0", sg)
	}
}

func TestSoftmaxSumsToOne(t *testing.T) {
	p := Softmax(Vec{2, 1, 0.1, -3})
	var s float64
	for _, v := range p {
		s += v
	}
	if !approx(s, 1, 1e-9) {
		t.Fatalf("softmax sum = %v, want 1", s)
	}
}

func TestCrossEntropyGrad(t *testing.T) {
	p := Softmax(Vec{1, 2, 3})
	loss, grad := CrossEntropy(p, 2)
	if loss <= 0 {
		t.Fatalf("loss = %v, want > 0", loss)
	}
	// grad should sum to ~0 (probs sum to 1, minus one-hot).
	var s float64
	for _, v := range grad {
		s += v
	}
	if !approx(s, 0, 1e-9) {
		t.Fatalf("grad sum = %v, want 0", s)
	}
}