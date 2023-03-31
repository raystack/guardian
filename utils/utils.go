package utils

func IsInteger(val float64) bool {
	return val == float64(int(val))
}
