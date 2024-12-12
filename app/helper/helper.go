package helper

func Abs(x int) int {
	mask := x >> 31
	return (x ^ mask) - mask
}

func Zero[T any]() T {
	var t T
	return t
}
