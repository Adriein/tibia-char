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
	if len(data) == 0 {
		return 0
	}

	sum := 0

	for _, value := range data {
		sum += value
	}

	return float64(sum) / float64(len(data))
}

func (s *Statistics) Variance(data []int) float64 {
	if len(data) < 2 {
		return 0
	}
	mean := s.Median(data)

	var sum float64

	for _, value := range data {
		diff := float64(value) - mean
		sum += diff * diff
	}

	return sum / float64(len(data)-1)
}

func (s *Statistics) StdDeviation(data []int) float64 {
	return math.Sqrt(s.Variance(data))
}

func (s *Statistics) Median(data []int) float64 {
	length := float64(len(data))

	dataCopyToSort := make([]int, len(data))

	copy(dataCopyToSort, data)

	sort.Ints(dataCopyToSort)

	if length == 0 {
		return 0
	}

	middle := int(length / 2)

	if len(dataCopyToSort)%2 == 1 {
		return float64(dataCopyToSort[middle])
	}

	return float64(dataCopyToSort[middle-1]+dataCopyToSort[middle]) / 2
}
