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
// extractTransactions собирает транзакции, используя Zip-подход для обхода колоночного текста
func extractTransactions(userID uuid.UUID, lines []string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	// Наши "корзины" для колонок
	var dates []string
	var amounts []string
	var categories []string
	var details []string

	dateRe := regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{2})$`)
	// Ловит любые суммы, игнорируя съехавшие минусы (напр. "18 - 850,00 T" или "0,00 T")
	amountRe := regexp.MustCompile(`^[\d\s\p{Z}+-]+,\d{2}`)

	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Игнорируем пустые строки и артефакты PDF-парсера (например, одиночные дефисы)
		if line == "" || line == "-" {
			continue
		}

		// Триггер начала таблицы: либо заголовок "Детали", либо первая найденная дата.
		// Это позволяет проигнорировать все сводные суммы из начала выписки.
		if line == "Детали" || dateRe.MatchString(line) {
			inTable = true
		}

		if !inTable {
			continue
		}

		// Игнорируем колонтитулы следующей страницы и системный текст
		if isKaspiSystemText(line) {
			continue
		}

		// Распределяем данные по корзинам
		if dateRe.MatchString(line) {
			dates = append(dates, line)
		} else if amountRe.MatchString(line) {
			amounts = append(amounts, line)
		} else if isKaspiCategory(line) {
			categories = append(categories, line)
		} else {
			// Всё, что не попало в регулярки выше — это детали транзакции (имена, магазины)
			details = append(details, line)
		}
	}

	// Находим минимальную длину, чтобы избежать паники, если парсер где-то проглотил строку
	minLen := len(dates)
	if len(amounts) < minLen { minLen = len(amounts) }
	if len(categories) < minLen { minLen = len(categories) }
	if len(details) < minLen { minLen = len(details) }

	// Оставляем логгирование для отладки
	if len(dates) != len(amounts) || len(amounts) != len(categories) || len(categories) != len(details) {
		fmt.Printf("Warning: Column length mismatch! Dates: %d, Amounts: %d, Categories: %d, Details: %d\n",
			len(dates), len(amounts), len(categories), len(details))
	}

	// Сшиваем транзакции вместе (Zip)
	for i := 0; i < minLen; i++ {
		amount := parseKaspiAmount(amounts[i])

		tType := domain.Expense
		if amount > 0 {
			tType = domain.Income
		}

		// В БД всегда сохраняем сумму как положительную
		if amount < 0 {
			amount = -amount
		}

		dateParsed, err := time.Parse("02.01.06", dates[i])
		if err != nil {
			dateParsed = time.Now()
		}

		tx := &domain.Transaction{
			ID:          uuid.New(),
			UserID:      userID,
			Amount:      amount,
			Description: details[i],
			Category:    categories[i],
			Type:        tType,
			Status:      domain.StatusCategorized,
			CreatedAt:   dateParsed,
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// parseKaspiAmount стала "пуленепробиваемой"
func parseKaspiAmount(s string) float64 {
	sign := 1.0
	// Если минус есть где угодно (даже если он съехал внутрь: "18 - 850") — это трата
	if strings.Contains(s, "-") {
		sign = -1.0
	}

	// Жестко вытаскиваем только цифры и запятую (убивает "T", "₸", "〒", любые пробелы и буквы)
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

// Вспомогательная функция проверки категорий
func isKaspiCategory(s string) bool {
	// Каспи использует ограниченный пул категорий
	categories := []string{"Пополнение", "Перевод", "Покупка", "Снятие", "Вознаграждение", "Комиссия"}
	for _, c := range categories {
		if s == c {
			return true
		}
	}
	return false
}

// Вспомогательная функция для отсева футеров и шапок страниц
func isKaspiSystemText(s string) bool {
	lower := strings.ToLower(s)

	exactMatches := []string{
		"дата", "сумма", "операция", "детали",
	}
	for _, text := range exactMatches {
		if lower == text {
			return true
		}
	}

	// Отсеиваем служебные подписи снизу страниц
	if strings.Contains(lower, "kaspi bank") ||
		strings.Contains(lower, "бик caspkzka") ||
		strings.Contains(lower, "www.kaspi.kz") ||
		strings.Contains(lower, "сумма заблокирована") ||
		strings.Contains(lower, "ожидает подтверждения") {
		return true
	}

	return false
}