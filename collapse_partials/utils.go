package main

// Calculate absolute value of integer.
func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func MaxTwoInts(a, b int) int {
	if a > b {
		return a
	}
	return b
}
