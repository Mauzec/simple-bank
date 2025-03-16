package util

func IsInt(val float64) bool {
	return val == float64(int(val))
}
