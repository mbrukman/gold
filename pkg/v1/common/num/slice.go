package num

import (
	"math/rand"
)

// MakeIRange creates an int slice.
func MakeIRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

// RandF32 returns a random float in the given range.
func RandF32(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

// MaxF32 returns the max index and max value for a float32 slice.
func MaxF32(vals []float32) (int, float32) {
	var max float32
	var maxI int
	for i, v := range vals {
		if i == 0 || v > max {
			max = v
			maxI = i
		}
	}
	return maxI, max
}

// I32SliceToI converts an int32 slice to an int slice.
func I32SliceToI(s []int32) []int {
	ret := []int{}
	for _, i := range s {
		ret = append(ret, int(i))
	}
	return ret
}
