package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// ParserUseCase handles parsing bank statements and saving transactions
type ParserUseCase struct {
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
}

func NewParserUseCase(transactionRepo repository.TransactionRepository, userRepo repository.UserRepository) *ParserUseCase {
	return &ParserUseCase{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// UploadStatement parses a Kaspi PDF and creates all found transactions for the user
func (uc *ParserUseCase) UploadStatement(ctx context.Context, userID uuid.UUID, filePath string) ([]*domain.Transaction, error) {
	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Parse the PDF
	transactions, err := uc.parseKaspiPDF(userID, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Save all parsed transactions to the database
	for _, tx := range transactions {
		if err := uc.transactionRepo.Create(ctx, tx); err != nil {
			fmt.Printf("Warning: failed to save transaction %s: %v\n", tx.ID, err)
		}
	}

	return transactions, nil
}

// parseKaspiPDF extracts text from a Kaspi bank statement PDF and returns parsed transactions
func (uc *ParserUseCase) parseKaspiPDF(userID uuid.UUID, filePath string) ([]*domain.Transaction, error) {
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
	return extractTransactions(userID, lines)
}

// extractTransactions processes lines from the PDF and builds Transaction objects.
// Kaspi PDF format puts each field on a separate line:
//   Line 1: Date (DD.MM.YY)
//   Line 2: Amount (e.g. "- 40 000,00 ₸")
//   Line 3: Operation type (Перевод, Покупка, Пополнение, Снятие)
//   Line 4: Details (На Kaspi Депозит, Ясмін, etc.)
func extractTransactions(userID uuid.UUID, lines []string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	dateRe := regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})$`)
	amountRe := regexp.MustCompile(`^([+-]?\s*[\d\s]+,\d{2})\s*₸$`)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Look for a date line, then read the next 3 lines
		if dateRe.MatchString(line) {
			dateStr := line

			if i+3 < len(lines) {
				amountLine := strings.TrimSpace(lines[i+1])
				operationLine := strings.TrimSpace(lines[i+2])
				detailsLine := strings.TrimSpace(lines[i+3])

				if amountRe.MatchString(amountLine) {
					amount := parseKaspiAmount(amountLine)

					tType := domain.Expense
					if amount > 0 {
						tType = domain.Income
					}

					// Always store amount as positive — Type field determines direction
					if amount < 0 {
						amount = -amount
					}

					dateParsed, err := time.Parse("02.01.06", dateStr)
					if err != nil {
						dateParsed = time.Now()
					}

					tx := &domain.Transaction{
						ID:          uuid.New(),
						UserID:      userID,
						Amount:      amount,
						Description: detailsLine,
						Category:    operationLine,
						Type:        tType,
						Status:      domain.StatusCategorized,
						CreatedAt:   dateParsed,
					}

					transactions = append(transactions, tx)

					// Skip the 3 lines we just consumed
					i += 3
					continue
				}
			}
		}
	}

	return transactions, nil
}

// parseKaspiAmount converts Kaspi-formatted amount strings to float64.
// Handles non-breaking spaces (U+00A0) that PDFs use between digit groups.
// Examples: "- 40 000,00 ₸" -> -40000.00, "+ 5 000,00 ₸" -> 5000.00
func parseKaspiAmount(s string) float64 {
	// Strip the ₸ sign
	s = strings.ReplaceAll(s, "₸", "")
	s = strings.TrimSpace(s)

	sign := 1.0
	if strings.HasPrefix(s, "-") {
		sign = -1.0
		s = strings.TrimPrefix(s, "-")
	} else if strings.HasPrefix(s, "+") {
		s = strings.TrimPrefix(s, "+")
	}

	// Strip ALL unicode whitespace (including non-breaking space U+00A0 from PDFs)
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\u00A0' || r == '\u202F' || r == '\u2009' {
			return -1
		}
		return r
	}, s)

	// Comma -> period for float conversion
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return sign * val
}
