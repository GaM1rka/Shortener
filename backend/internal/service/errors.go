package service

import "errors"

var (
	ErrInvalidURL          = errors.New("invalid url")
	ErrShortCodeNotFound   = errors.New("short code not found")
	ErrShortCodeExists     = errors.New("short code already exists")
	ErrInvalidCustomAlias  = errors.New("custom alias must be 3-64 chars and contain only letters, numbers, underscores, and dashes")
	ErrUnableToCreateAlias = errors.New("unable to create unique short code")
)
