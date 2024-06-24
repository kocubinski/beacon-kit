package sszdb

import (
	"math"
)

func floorLog2(n uint64) uint64 {
	return uint64(math.Floor(math.Log2(float64(n))))
}

func ceilLog2(n uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(n))))
}

func powerTwo(n uint64) uint64 {
	return uint64(math.Pow(2, float64(n)))
}

func nextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return uint64(v)
}

func prevPowerOfTwo(v uint64) uint64 {
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return uint64(v) - (v >> 1)
}
