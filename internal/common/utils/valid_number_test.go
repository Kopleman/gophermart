package utils

import "testing"

func TestIsValidOrderNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Валидные номера
		{"Valid classic Luhn", "79927398713", true},
		{"Valid Visa-like number", "49927398716", true},
		{"Valid short number", "12345674", true},
		{"Valid 16-digit number", "4242424242424242", true},
		{"Valid zero number", "0", true},
		{"Valid double zero", "00", true},
		{"Valid number with zero", "1234567812345670", true},

		// Невалидные номера
		{"Invalid check digit", "79927398712", false},
		{"Invalid short number", "123", false},
		{"Invalid character in string", "4992a7398716", false},
		{"Empty string", "", false},
		{"Invalid modified number", "4242424242424241", false},
		{"Invalid long number", "1234567812345678", false},
		{"Invalid with symbols", "4111-1111-1111-1111", false},
		{"Invalid with spaces", "4992 7398 716", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidOrderNumber(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidOrderNumber(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	// Тест на обработку чисел с переносом через 9
	t.Run("Sum overflow handling", func(t *testing.T) {
		// 6*2 = 12 → 1+2=3, 8 → 3+8=11 → 11%10 !=0
		if IsValidOrderNumber("68") {
			t.Error("Expected invalid for '68'")
		}
	})

	// Тест минимального валидного числа
	t.Run("Minimal valid number", func(t *testing.T) {
		if !IsValidOrderNumber("0") {
			t.Error("Expected valid for '0'")
		}
	})

	// Тест нечетной длины
	t.Run("Odd length valid", func(t *testing.T) {
		if !IsValidOrderNumber("12345674") {
			t.Error("Expected valid for '12345674'")
		}
	})
}
