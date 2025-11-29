package statistics

import (
	"math"
	"sort"
)

type Statistics struct{}

func New() *Statistics {
	return &Statistics{}
}

func (s *Statistics) Mean(data []int) float64 {
	sum := 0

	for _, value := range data {
		sum += value
	}

	return float64(sum) / float64(len(data))
}

func (s *Statistics) Variance(data []int) float64 {
	mean := int(s.Mean(data))
	sum := 0

	for _, value := range data {
		diff := value - mean
		sum += diff * diff
	}
	return float64(sum) / float64(len(data)-1)
}

func (s *Statistics) StdDeviation(data []int) float64 {
	return math.Sqrt(s.Variance(data))
}

func (s *Statistics) Median(data []int) int {
	length := float64(len(data))

	dataCopyToSort := make([]int, len(data))

	copy(dataCopyToSort, data)

	sort.Ints(dataCopyToSort)

	if length == 0 {
		return 0
	}

	middle := int(length / 2)

	if len(dataCopyToSort)%2 == 1 {
		return dataCopyToSort[middle]
	}

	return int((float64(dataCopyToSort[middle-1] + dataCopyToSort[middle])) / 2)
}
