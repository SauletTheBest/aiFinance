package kaspi

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// RawTransaction holds parsed data from a Kaspi PDF statement
// without any domain-specific types or dependencies.
type RawTransaction struct {
	Date        time.Time
	Amount      float64 // negative = expense, positive = income
	Category    string
	Description string
}

// ParsePDF reads a Kaspi bank statement PDF and returns raw parsed transactions.
func ParsePDF(filePath string) ([]RawTransaction, error) {
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
	lines := strings.Split(text, "\n")
	return extractTransactions(lines)
}

// extractTransactions processes lines from the PDF and builds RawTransaction slices
// using a Zip-approach for the columnar text layout of Kaspi statements.
//
// Kaspi PDF column order:
//   - Column 1: Date (DD.MM.YY)
//   - Column 2: Amount (e.g. "- 40 000,00 ₸")
//   - Column 3: Operation type (Перевод, Покупка, Пополнение, Снятие)
//   - Column 4: Details (На Kaspi Депозит, Ясмін, etc.)
func extractTransactions(lines []string) ([]RawTransaction, error) {
	var dates []string
	var amounts []string
	var categories []string
	var details []string

	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and PDF artifacts (standalone dashes)
		if line == "" || line == "-" {
			continue
		}

		// Trigger table start on header "Детали" or first date
		if line == "Детали" || IsDate(line) {
			inTable = true
		}

		if !inTable {
			continue
		}

		// Skip page headers/footers
		if IsSystemText(line) {
			continue
		}

		// Skip foreign currency notes like "(- 5,80 USD)" — these are annotations, not transactions
		if currencyNoteRe.MatchString(line) {
			continue
		}

		// Skip standalone plus signs or currency symbols
		if line == "+" || line == "₸" || line == "T" {
			continue
		}

		// Route each line into the appropriate column bucket
		switch {
		case IsDate(line):
			dates = append(dates, line)
		case IsAmount(line):
			amounts = append(amounts, line)
		case IsCategory(line):
			categories = append(categories, line)
		default:
			details = append(details, line)
		}
	}

	// Find minimum length to avoid panic if the parser missed a line
	minLen := len(dates)
	if len(amounts) < minLen {
		minLen = len(amounts)
	}
	if len(categories) < minLen {
		minLen = len(categories)
	}
	if len(details) < minLen {
		minLen = len(details)
	}

	// Log column mismatch for debugging (not returned as error since partial results are still useful)
	if len(dates) != len(amounts) || len(amounts) != len(categories) || len(categories) != len(details) {
		log.Printf("Warning: column length mismatch — Dates: %d, Amounts: %d, Categories: %d, Details: %d",
			len(dates), len(amounts), len(categories), len(details))
	}

	// Zip columns together into raw transactions
	transactions := make([]RawTransaction, 0, minLen)
	for i := 0; i < minLen; i++ {
		dateParsed, err := time.Parse("02.01.06", dates[i])
		if err != nil {
			dateParsed = time.Now()
		}

		transactions = append(transactions, RawTransaction{
			Date:        dateParsed,
			Amount:      ParseAmount(amounts[i]),
			Category:    categories[i],
			Description: details[i],
		})
	}

	return transactions, nil
}
