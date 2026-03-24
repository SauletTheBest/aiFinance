package validator

import (
	"net/mail"
	"regexp" // можно потом добавить полноценную валидацию
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidateRegistration(name, email, password, currency string) []*ValidationError {
	var errors []*ValidationError

	// Name validation
	if name == "" {
		errors = append(errors, &ValidationError{Field: "name", Message: "name is required"})
	} else if len(name) < 2 {
		errors = append(errors, &ValidationError{Field: "name", Message: "name must be at least 2 characters"})
	} else if len(name) > 100 {
		errors = append(errors, &ValidationError{Field: "name", Message: "name must be less than 100 characters"})
	}

	// Email validation
	if email == "" {
		errors = append(errors, &ValidationError{Field: "email", Message: "email is required"})
	} else if _, err := mail.ParseAddress(email); err != nil {
		errors = append(errors, &ValidationError{Field: "email", Message: "invalid email format"})
	}

	// Password validation
	if password == "" {
		errors = append(errors, &ValidationError{Field: "password", Message: "password is required"})
	} else if len(password) < 8 {
		errors = append(errors, &ValidationError{Field: "password", Message: "password must be at least 8 characters"})
	} else if len(password) > 128 {
		errors = append(errors, &ValidationError{Field: "password", Message: "password must be less than 128 characters"})
	}
	//Currency Validation
	if currency == "" {
		errors = append(errors, &ValidationError{Field: "currency", Message: "currency is required"})
	} else if len(currency) > 3 {
		errors = append(errors, &ValidationError{Field: "currency", Message: "currency name is too long"})
	} else if len(currency) < 3 {
		errors = append(errors, &ValidationError{Field: "currency", Message: "currency name is too short"})
	} else if !regexp.MustCompile(`^[A-Z]+$`).MatchString(currency) {
		errors = append(errors, &ValidationError{Field: "currency", Message: "currency should be in uppercase"})
	}

	return errors
}

func ValidateTransaction(amount float64, description string, category string) []*ValidationError {
	var errors []*ValidationError

	// Amount validation
	if amount == 0 {
		errors = append(errors, &ValidationError{Field: "amount", Message: "amount cannot be zero"})
	} else if amount < -999999999 {
		errors = append(errors, &ValidationError{Field: "amount", Message: "expense amount is too large"})
	} else if amount > 999999999 {
		errors = append(errors, &ValidationError{Field: "amount", Message: "income amount is too large"})
	}

	// Description validation
	if description == "" {
		errors = append(errors, &ValidationError{Field: "description", Message: "description is required"})
	} else if len(description) > 500 {
		errors = append(errors, &ValidationError{Field: "description", Message: "description must be less than 500 characters"})
	}

	// category
	if category == "" {
		errors = append(errors, &ValidationError{Field: "category", Message: "category is required"})
	} else if len(category) > 30 {
		errors = append(errors, &ValidationError{Field: "category", Message: "category must be less than 30 characters"})
	}

	return errors
}

func ValidateUpdateTransaction(amount *float64, description *string, category *string) []*ValidationError {
	var errors []*ValidationError

	// Amount validation (if provided)
	if amount != nil {
		if *amount == 0 {
			errors = append(errors, &ValidationError{Field: "amount", Message: "amount cannot be zero"})
		} else if *amount < -999999999 {
			errors = append(errors, &ValidationError{Field: "amount", Message: "expense amount is too large"})
		} else if *amount > 999999999 {
			errors = append(errors, &ValidationError{Field: "amount", Message: "income amount is too large"})
		}
	}

	// Description validation (if provided)
	if description != nil && *description == "" {
		errors = append(errors, &ValidationError{Field: "description", Message: "description cannot be empty"})
	} else if description != nil && len(*description) > 500 {
		errors = append(errors, &ValidationError{Field: "description", Message: "description must be less than 500 characters"})
	}

	//category
	if category != nil && *category == "" {
		errors = append(errors, &ValidationError{Field: "category", Message: "category cannot be empty"})
	} else if category != nil && len(*category) > 30 {
		errors = append(errors, &ValidationError{Field: "category", Message: "category must be less than 30 characters"})
	}

	return errors
}
