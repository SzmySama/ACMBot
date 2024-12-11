package helper

func Abs(x int) int {
	mask := x >> 31
	return (x ^ mask) - mask
}
