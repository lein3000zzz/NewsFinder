package nlp

import "math"

func convertIntToInt64(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}

func softmax(logits []float32) []float32 {
	var sum float64
	res := make([]float32, len(logits))
	for i, v := range logits {
		res[i] = float32(math.Exp(float64(v)))
		sum += float64(res[i])
	}
	for i := range res {
		res[i] /= float32(sum)
	}
	return res
}
