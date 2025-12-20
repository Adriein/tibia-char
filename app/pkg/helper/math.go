package helper

func PercentSafe(numerator interface{}, denominator interface{}) float64 {
	var num, den float64

	switch v := numerator.(type) {
	case int:
		num = float64(v)
	case float64:
		num = v
	}

	switch v := denominator.(type) {
	case int:
		den = float64(v)
	case float64:
		den = v
	}

	if den == 0 {
		return 0
	}

	return (num / den) * 100
}
