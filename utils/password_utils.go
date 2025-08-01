package utils

import (
	"regexp"
	"unicode"
)

func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(string(char)):
			hasSpecial = true
		}
	}

	// 要求至少包含大写字母、小写字母、数字和特殊字符中的三种
	criteriaMet := 0
	if hasUpper {
		criteriaMet++
	}
	if hasLower {
		criteriaMet++
	}
	if hasNumber {
		criteriaMet++
	}
	if hasSpecial {
		criteriaMet++
	}

	return criteriaMet >= 3
}
