package kaspi

import (
	"regexp"
	"strconv"
	"strings"
)

// dateRe matches Kaspi statement dates in DD.MM.YY format (strictly standalone)
var dateRe = regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})$`)

// Improved amountRe according to your suggestion
var amountRe = regexp.MustCompile(`^[+-]?\s*\d[\d\s]*,\d{2}`)

// Компилируем регулярное выражение ОДИН РАЗ на уровне пакета, чтобы ParseAmount не тормозил в цикле
var amountDigitsRe = regexp.MustCompile(`[\d\s]+,\d{2}`)

// Категории упорядочены от самых длинных/специфичных к самым коротким/общим!
var kaspiCategories = []string{
	"Поступление со своего счета",
	"Поступление",
	"Перевод на свой счет",
	"Перевод на свой",
	"Перевод",
	"Пополнение",
	"Покупка",
	"Снятие",
	"Вознаграждение",
	"Комиссия",
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

// ParseAmount извлекает float64 из строки суммы Kaspi с помощью скомпилированной регулярки.
func ParseAmount(s string) float64 {
	clean := strings.TrimSpace(s)
	negative := strings.HasPrefix(clean, "-")

	// Используем уже скомпилированную глобально регулярку
	num := amountDigitsRe.FindString(clean)

	num = strings.ReplaceAll(num, " ", "")
	num = strings.ReplaceAll(num, ",", ".")

	val, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0
	}

	if negative {
		return -val
	}
	return val
}
