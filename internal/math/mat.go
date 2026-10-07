package math

import (
	"runtime"
	"sync"
)

// Mat is a row-major dense matrix (Rows x Cols).
type Mat struct {
	Rows, Cols int
	Data       Vec // len == Rows*Cols
}

// NewMat allocates a zeroed Rows x Cols matrix.
func NewMat(rows, cols int) *Mat {
	return &Mat{Rows: rows, Cols: cols, Data: make(Vec, rows*cols)}
}

// At returns element (r,c).
func (m *Mat) At(r, c int) float64 { return m.Data[r*m.Cols+c] }

// Set assigns element (r,c).
func (m *Mat) Set(r, c int, v float64) { m.Data[r*m.Cols+c] = v }

// Row returns a view of row r (shares backing storage).
func (m *Mat) Row(r int) Vec { return m.Data[r*m.Cols : (r+1)*m.Cols] }

// MatVec computes y = m * x  (len(x)==Cols, len(y)==Rows).
func (m *Mat) MatVec(x Vec) Vec {
	if len(x) != m.Cols {
		panic("bones: MatVec dimension mismatch")
	}
	y := make(Vec, m.Rows)
	for r := 0; r < m.Rows; r++ {
		y[r] = Dot(m.Row(r), x)
	}
	return y
}

// MatMul computes C = A * B, parallelised across rows of A using all cores.
// This is the hot path for the cerebellum's forward/backward passes.
func MatMul(a, b *Mat) *Mat {
	if a.Cols != b.Rows {
		panic("bones: MatMul dimension mismatch")
	}
	c := NewMat(a.Rows, b.Cols)
	workers := runtime.GOMAXPROCS(0)
	if workers > a.Rows {
		workers = a.Rows
	}
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	chunk := (a.Rows + workers - 1) / workers
	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > a.Rows {
			hi = a.Rows
		}
		if lo >= hi {
			break
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				arow := a.Row(i)
				crow := c.Row(i)
				for k := 0; k < a.Cols; k++ {
					av := arow[k]
					if av == 0 {
						continue
					}
					brow := b.Row(k)
					AXPY(av, brow, crow)
				}
			}
		}(lo, hi)
	}
	wg.Wait()
	return c
}

// Transpose returns a new matrix that is m^T.
func (m *Mat) Transpose() *Mat {
	t := NewMat(m.Cols, m.Rows)
	for r := 0; r < m.Rows; r++ {
		for c := 0; c < m.Cols; c++ {
			t.Set(c, r, m.At(r, c))
		}
	}
	return t
}