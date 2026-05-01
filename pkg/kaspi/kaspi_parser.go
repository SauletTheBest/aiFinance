package kaspi

import (
	"fmt"
	"strings"
	"time"
	"log"
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
	var transactions []RawTransaction
	var current *RawTransaction
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

		// 💡 A new date means a NEW transaction row starts!
		if IsDate(line) {
			if current != nil {
				// Log a warning if we are saving a transaction that has no amount
				if current.Amount == 0 {
					log.Printf("Warning: Transaction on %s has no amount. Details: %s", current.Date.Format("02.01.06"), current.Description)
				}
				transactions = append(transactions, *current)
			}
			dateParsed, err := time.Parse("02.01.06", line)
			if err != nil {
				dateParsed = time.Now()
			}
			current = &RawTransaction{Date: dateParsed}
			continue
		}

		if current == nil {
			continue
		}

		// 💡 Much cleaner switch case!
		switch {
		case IsAmount(line) && current.Amount == 0:
			current.Amount = ParseAmount(line)
			
		case IsCategory(line) && current.Category == "":
			current.Category = line
			
		default:
			// If it's not a date, amount, or category, it MUST be a detail.
			if current.Description == "" {
				current.Description = line
			} else {
				current.Description += " " + line // Appends multi-line details safely!
			}
		}
	}

	// Don't forget to add the very last transaction!
	if current != nil {
		transactions = append(transactions, *current)
	}

	// 💡 Great success log!
	log.Printf("Successfully extracted %d transactions using Row-Based parsing.", len(transactions))

	return transactions, nil
}
