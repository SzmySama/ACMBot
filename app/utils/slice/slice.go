package slice

func Reverse[T any](slice *[]T) {
	left, right := 0, len(*slice)-1
	for left < right {
		(*slice)[left], (*slice)[right] = (*slice)[right], (*slice)[left]
		left++
		right--
	}
}
