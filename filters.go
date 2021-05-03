package main

import (
	"time"

	"github.com/flosch/pongo2"
	"github.com/shopspring/decimal"
)

// Date is a drop-in replacement for the standard "date"-filter, but with support for time.Time, *time.Time and nil.
// It also has a default value set, so no parameter is necessary
func Date(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	var t *time.Time
	isTime := true
	switch v := in.Interface().(type) {
	case *time.Time:
		t = v
	case time.Time:
		t = &v
	default:
		isTime = false
	}

	if !isTime || t == nil || t.IsZero() {
		return pongo2.AsValue("-"), nil
	}

	format := "2006-01-02"
	if param != nil && param.String() != "" {
		format = param.String()
	}

	return pongo2.AsValue(t.Format(format)), nil
}

// Money converts a decimal.Decimal value to a 2-point fixed string
func Money(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	moneyVal := "-"
	switch v := in.Interface().(type) {
	case *decimal.Decimal:
		moneyVal = v.StringFixedBank(2)
	case decimal.Decimal:
		moneyVal = v.StringFixedBank(2)
	case decimal.NullDecimal:
		if v.Valid {
			moneyVal = v.Decimal.StringFixedBank(2)
		}
	}

	return pongo2.AsValue(moneyVal), nil
}
