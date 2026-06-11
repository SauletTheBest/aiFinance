package kaspi

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// ParsedStatement holds parsed transactions and balance details from Kaspi statement
type ParsedStatement struct {
	Transactions    []RawTransaction
	StartingBalance float64
	FinalBalance    float64
}

// RawTransaction holds parsed data from a Kaspi PDF statement
type RawTransaction struct {
	Date        time.Time
	Amount      float64 // negative = expense, positive = income
	Category    string
	Description string
}

// Регулярные выражения скомпилированы один раз на уровне пакета
var (
	datePrefixRe  = regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})`)
	transactionRe = regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})\s+([+-]?\s*[\d\s]+,\d{2})\s*(?:₸|T|тнг|KZT|тг)?\s+(.+)$`)
	
	// Регулярные выражения для поиска стартового и финального баланса
	startingBalanceRe = regexp.MustCompile(`Доступно на \d{2}\.\d{2}\.\d{2}\s+([+-]?\s*[\d\s]+,\d{2})\s*(?:₸|T|тнг|KZT|тг)?`)
	finalBalanceRe    = regexp.MustCompile(`Доступно на \d{2}\.\d{2}\.\d{2}:\s*([+-]?\s*[\d\s]+,\d{2})\s*(?:₸|T|тнг|KZT|тг)?`)
)

// ParsePDF reads a Kaspi bank statement PDF and returns parsed statement (transactions + balances).
func ParsePDF(filePath string) (*ParsedStatement, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf: %w", err)
	}
	defer f.Close()

	var contentBuilder strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err == nil {
			contentBuilder.WriteString(text)
			contentBuilder.WriteString("\n")
		}
	}

	text := contentBuilder.String()

	// Находим стартовый и финальный балансы на основе полного текста выписки
	var startingBalance float64
	var finalBalance float64

	if matches := startingBalanceRe.FindStringSubmatch(text); len(matches) == 2 {
		startingBalance = ParseAmount(matches[1])
	}
	if matches := finalBalanceRe.FindStringSubmatch(text); len(matches) == 2 {
		finalBalance = ParseAmount(matches[1])
	}

	lines := strings.Split(text, "\n")
	transactions, err := extractTransactions(lines)
	if err != nil {
		return nil, err
	}

	return &ParsedStatement{
		Transactions:    transactions,
		StartingBalance: startingBalance,
		FinalBalance:    finalBalance,
	}, nil
}

// NormalizeLines склеивает перенесенные строки (описания) с основной транзакцией,
// при этом корректно игнорирует системный текст, чтобы он не приклеился к транзакции.
func NormalizeLines(lines []string) []string {
	var result []string
	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "-" {
			continue
		}

		// Таблица транзакций начинается с первой строки с датой
		if IsDatePrefix(line) {
			inTable = true
		}

		if !inTable {
			continue
		}

		// Игнорируем системные строки (шапки страниц, футеры)
		if IsSystemText(line) {
			continue
		}

		if IsDatePrefix(line) {
			result = append(result, line)
		} else if len(result) > 0 {
			// Если строка не начинается с даты, значит это продолжение описания предыдущей транзакции!
			result[len(result)-1] += " " + line
		}
	}

	return result
}

// IsDatePrefix проверяет, начинается ли строка с даты формата DD.MM.YY
func IsDatePrefix(s string) bool {
	return datePrefixRe.MatchString(s)
}

// SplitCategoryAndDescription находит категорию из списка Kaspi и отделяет ее от описания
func SplitCategoryAndDescription(s string) (string, string) {
	for _, cat := range kaspiCategories {
		if strings.HasPrefix(s, cat) {
			desc := strings.TrimSpace(strings.TrimPrefix(s, cat))
			return cat, desc
		}
	}
	return "", s
}

// extractTransactions обрабатывает строки и преобразует их в транзакции
func extractTransactions(lines []string) ([]RawTransaction, error) {
	normalizedLines := NormalizeLines(lines)
	var transactions []RawTransaction

	for _, line := range normalizedLines {
		matches := transactionRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			dateStr := matches[1]
			amountStr := matches[2]
			rest := matches[3]

			category, description := SplitCategoryAndDescription(rest)

			dateParsed, err := time.Parse("02.01.06", dateStr)
			if err != nil {
				dateParsed = time.Now()
			}

			transactions = append(transactions, RawTransaction{
				Date:        dateParsed,
				Amount:      ParseAmount(amountStr),
				Category:    category,
				Description: description,
			})
		} else {
			// Логируем неразобранные строки для последующего анализа
			log.Printf("UNPARSED: %s", line)
		}
	}

	log.Printf("Successfully extracted %d transactions using Hybrid (Row-Based) parsing.", len(transactions))
	return transactions, nil
}
