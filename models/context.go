package models

type contextKey int

const (
	SessionContextKey contextKey = iota
	TransactionContextKey
)
