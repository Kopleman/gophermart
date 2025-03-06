package utils

func IsValidOrderNumber(numberStr string) bool {
	if len(numberStr) == 0 {
		return false
	}

	for _, c := range numberStr {
		if c < '0' || c > '9' {
			return false
		}
	}

	digits := make([]int, len(numberStr))
	for i, c := range numberStr {
		digits[i] = int(c - '0')
	}

	sum := 0
	double := false

	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]

		if double {
			digit *= 2
			if digit > 9 { //nolint:all // too tedious to move it
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
