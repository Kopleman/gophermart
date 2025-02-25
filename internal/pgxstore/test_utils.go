package pgxstore

import "github.com/shopspring/decimal"

type TestDecimal struct {
	expected decimal.Decimal
}

func (dt TestDecimal) Match(v interface{}) bool {
	d, ok := v.(decimal.Decimal)
	if !ok {
		return false
	}
	return d.Equal(dt.expected)
}
