package utils

func CalcularPaginas(total, size int32) int32 {
	if size <= 0 {
		return 0
	}
	return (total + size - 1) / size
}
