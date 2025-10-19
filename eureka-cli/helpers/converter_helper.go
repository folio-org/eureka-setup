package helpers

func ConvertMiBToBytes(n int64) int64 {
	if n > 0 {
		return n * 1024 * 1024
	}
	return n
}
