// Package validation provides input validation helpers.
package validation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	// ErrEmptyLogin is returned when login is blank.
	ErrEmptyLogin = errors.New("login is required")
	// ErrEmptyPassword is returned when password is blank.
	ErrEmptyPassword = errors.New("password is required")
	// ErrEmptyName is returned when secret name is blank.
	ErrEmptyName = errors.New("name is required")
	// ErrInvalidCard is returned when a card number fails Luhn check.
	ErrInvalidCard = errors.New("invalid card number")
)

// Credentials validates login and password for registration or authentication.
func Credentials(login, password string) error {
	if strings.TrimSpace(login) == "" {
		return ErrEmptyLogin
	}
	if password == "" {
		return ErrEmptyPassword
	}
	return nil
}

// SecretName validates a non-empty secret name.
func SecretName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyName
	}
	return nil
}

// Luhn validates a card number using the Luhn algorithm.
func Luhn(number string) error {
	digits := make([]int, 0, len(number))
	for _, r := range number {
		if r < '0' || r > '9' {
			continue
		}
		digits = append(digits, int(r-'0'))
	}
	if len(digits) < 13 || len(digits) > 19 {
		return ErrInvalidCard
	}
	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		n := digits[i]
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	if sum%10 != 0 {
		return ErrInvalidCard
	}
	return nil
}

// NonEmptyUTF8 reports whether s contains non-whitespace UTF-8 text.
func NonEmptyUTF8(s string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(s)) > 0
}
