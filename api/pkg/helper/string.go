package helper

import "unicode"

func CamelToSnake(str string) string {
	var result []rune

	for i, char := range str {
		if unicode.IsUpper(char) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(char))
	}

	return string(result)
}

func SnakeToCamel(str string) string {
	var result []rune
	capitalizeNext := false
	for _, char := range str {
		if char == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result = append(result, unicode.ToUpper(char))
			capitalizeNext = false
		} else {
			result = append(result, char)
		}
	}
	return string(result)
}
