package math

import "math"

// Activations and their derivatives for the cerebellum (the learned net).

// ReLU in place.
func ReLU(v Vec) {
	for i := range v {
		if v[i] < 0 {
			v[i] = 0
		}
	}
}

// ReLUGrad returns the gradient mask for a pre-activation vector.
func ReLUGrad(pre Vec) Vec {
	g := make(Vec, len(pre))
	for i := range pre {
		if pre[i] > 0 {
			g[i] = 1
		}
	}
	return g
}

// Tanh in place.
func Tanh(v Vec) {
	for i := range v {
		v[i] = math.Tanh(v[i])
	}
}

// Sigmoid of a scalar.
func Sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// Softmax returns a numerically stable softmax distribution.
func Softmax(v Vec) Vec {
	if len(v) == 0 {
		return Vec{}
	}
	max := v[0]
	for _, x := range v {
		if x > max {
			max = x
		}
	}
	out := make(Vec, len(v))
	var sum float64
	for i, x := range v {
		e := math.Exp(x - max)
		out[i] = e
		sum += e
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

// CrossEntropy returns the loss of a probability distribution against a target
// class index, plus the gradient w.r.t. the logits (probs - onehot).
func CrossEntropy(probs Vec, target int) (loss float64, grad Vec) {
	const eps = 1e-12
	loss = -math.Log(probs[target] + eps)
	grad = append(Vec(nil), probs...)
	grad[target] -= 1
	return
}