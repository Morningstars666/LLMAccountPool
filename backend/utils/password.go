package utils

import (
	"regexp"
)

// ValidatePasswordStrength 验证密码强度
// 要求：至少8个字符，至少一个大写字母，至少一个小写字母，至少一个数字，至少一个特殊字符
func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "密码长度至少8个字符"
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	if !hasUpper {
		return false, "密码必须包含至少一个大写字母"
	}
	if !hasLower {
		return false, "密码必须包含至少一个小写字母"
	}
	if !hasDigit {
		return false, "密码必须包含至少一个数字"
	}
	if !hasSpecial {
		return false, "密码必须包含至少一个特殊字符"
	}

	return true, "密码强度符合要求"
}
