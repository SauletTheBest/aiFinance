package kaspi

import (
	"regexp"
	"strconv"
	"strings"
)

// dateRe matches Kaspi statement dates in DD.MM.YY format
var dateRe = regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})$`)

// amountRe matches amounts, including misaligned minus signs (e.g. "18 - 850,00 T" or "0,00 T")
var amountRe = regexp.MustCompile(`^[\d\s\p{Z}+-]+,\d{2}`)

// kaspiCategories is the limited pool of operation types used by Kaspi Bank
var kaspiCategories = []string{
	"Пополнение", "Перевод", "Покупка", "Снятие", "Вознаграждение", "Комиссия",
}

// systemExactMatches are table header labels to skip
var systemExactMatches = []string{
	"дата", "сумма", "операция", "детали",
}

// systemSubstrings are footer/header fragments to skip
var systemSubstrings = []string{
	"kaspi bank",
	"бик caspkzka",
	"www.kaspi.kz",
	"сумма заблокирована",
	"ожидает подтверждения",
	"приложение к справке",        
	"раздел «краткое содержание",
	"содержит информацию об операциях",
}

// IsDate checks whether the line is a Kaspi date (DD.MM.YY)
func IsDate(s string) bool {
	return dateRe.MatchString(s)
}

// IsAmount checks whether the line looks like a Kaspi amount
func IsAmount(s string) bool {
	return amountRe.MatchString(s)
}

// IsCategory checks whether the line is one of Kaspi's known operation types
func IsCategory(s string) bool {
	for _, c := range kaspiCategories {
		if s == c {
			return true
		}
	}
	return false
}

// IsSystemText returns true for page headers, footers, and table labels
func IsSystemText(s string) bool {
	lower := strings.ToLower(s)

	for _, text := range systemExactMatches {
		if lower == text {
			return true
		}
	}

	for _, sub := range systemSubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}

	return false
}

// ParseAmount extracts a float64 from a Kaspi amount string.
// Handles misaligned minus signs (e.g. "18 - 850,00 ₸") and various currency symbols.
// Negative values represent expenses, positive — income.
func ParseAmount(s string) float64 {
	sign := 1.0
	if strings.Contains(s, "-") {
		sign = -1.0
	}

	// Keep only digits and comma — strip currency symbols, spaces, letters
	var builder strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == ',' {
			builder.WriteRune(r)
		}
	}

	cleanStr := strings.ReplaceAll(builder.String(), ",", ".")

	val, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0
	}
	return sign * val
}
