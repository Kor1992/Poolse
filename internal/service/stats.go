package service

func CalculateUptime(total, successful int64) float64 {
	if total == 0 {
		return 0
	}
	res := float64(successful) / float64(total) * 100
	return res
}
