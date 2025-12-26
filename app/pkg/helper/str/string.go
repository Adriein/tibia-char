package str

import (
	"strings"
	"unicode"
)

// CamelToSnake converts "MyString" to "my_string"
func CamelToSnake(s string) string {
	var res strings.Builder
	res.Grow(len(s) + 2)

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				res.WriteByte('_')
			}
			res.WriteRune(unicode.ToLower(r))
		} else {
			res.WriteRune(r)
		}
	}
	return res.String()
}

// SnakeToCamel converts "my_string" to "myString"
func SnakeToCamel(s string) string {
	var res strings.Builder
	res.Grow(len(s))

	capitalizeNext := false
	for _, r := range s {
		if r == '_' {
			capitalizeNext = true

			continue
		}
		if capitalizeNext {
			res.WriteRune(unicode.ToUpper(r))

			capitalizeNext = false

			continue
		}

		res.WriteRune(r)
	}

	return res.String()
}
