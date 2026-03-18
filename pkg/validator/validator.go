package validator


import (
    "net/mail"
    //"regexp" // можно потом добавить полноценную валидацию
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}

func ValidateRegistration(name, email, password string) []*ValidationError {
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
    
    return errors
}
