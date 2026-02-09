package luckymoney

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrIDNotIssued       = errors.New("id not issued")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrInfoMismatch      = errors.New("info mismatch")
	ErrIDAlreadyDrawn    = errors.New("id already drawn")
	ErrAlreadyDrawn      = errors.New("already drawn")
	ErrPoolEmpty         = errors.New("pool empty")
)
