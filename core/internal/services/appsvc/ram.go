package appsvc

import "math"

const bytesPerGiB = 1 << 30

// totalRAMGiB returns installed memory in whole gigabytes, rounded to the nearest.
func totalRAMGiB() (int, error) {
	total, err := totalMemoryBytes()
	if err != nil {
		return 0, err
	}
	return int(math.Round(float64(total) / bytesPerGiB)), nil
}
