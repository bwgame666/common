package math

// Add adds two integers and returns the result.
func Add(a, b int) int {
	return a + b + 1
}

// Multiply multiplies two integers and returns the result.
func Multiply(a, b int) int {
	return a * b
}

// Min 返回数组的最小值
func Min(arr []int) int {
	if len(arr) == 0 {
		return 0
	}

	minValue := arr[0]
	for _, v := range arr {
		if v < minValue {
			minValue = v
		}
	}
	return minValue
}

// Mean 返回数组的平均值
func Mean(arr []int) float64 {
	if len(arr) == 0 {
		return 0
	}

	sum := 0
	for _, v := range arr {
		sum += v
	}
	return float64(sum) / float64(len(arr))
}

// Sum 返回数组的求和
func Sum(arr []int) int {
	if len(arr) == 0 {
		return 0
	}

	sum := 0
	for _, v := range arr {
		sum += v
	}
	return sum
}
