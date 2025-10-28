package helpers

type ConversionMode int

const (
	BytesToMib = iota + 1
	MibToBytes
)

func ConvertMemory(m ConversionMode, n int64) int64 {
	if n > 0 {
		switch m {
		case BytesToMib:
			return n / 1024 / 1024
		case MibToBytes:
			return n * 1024 * 1024
		}
	}

	return n
}
