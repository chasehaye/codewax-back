package validation

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    if err != nil {
        return "", err
    }
    return string(hashed), nil
}

func ComparePassword(hashedPassword, plainPassword string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func ValidatePassword(password string) error {
	var errs []error
	if len(password) < 8 {
		errs = append(errs, errors.New("password must be at least 8 characters long"))
	}
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errs = append(errs, errors.New("password must contain at least one uppercase letter"))
	}
	if !hasLower {
		errs = append(errs, errors.New("password must contain at least one lowercase letter"))
	}
	if !hasNumber {
		errs = append(errs, errors.New("password must contain at least one number"))
	}
	if !hasSpecial {
		errs = append(errs, errors.New("password must contain at least one special character"))
	}

	if len(errs) == 0 {
        return nil
    }

    return errors.Join(errs...)
}