package numbers

func DefaultInt(val int, fallback int) int {
	if val == 0 {
		return fallback
	}

	return val
}
