package model

func abs(x int) int {
	mask := x >> 31
	return (x ^ mask) - mask
}
