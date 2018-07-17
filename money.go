package fatzebra

import (
	"math"
	"strconv"
)

// AUD represents the Australian dollars currency.
type AUD int64

// Float64 returns the number of dollars the currency represents as a float.
// WARNING: You shouldn't be performing any form of
// arithmetic or transformation with currency in float form.
func (a AUD) Float64() float64 {
	return float64(a) / 100.0
}

// Cents returns the number of cents the currency represents.
func (a AUD) Cents() int64 {
	return int64(a)
}

// MarshalJSON implements the JSON unmarshaller interface.
func (a AUD) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(a), 10)), nil
}

// NewAUDFromCents returns the currency representation from a given number
// of cents.
func NewAUDFromCents(cents int64) AUD {
	return AUD(cents)
}

// NewAUDFromDollars returns AUD using the given amount of dollars
// that are float64. WARNING: You shouldn't be performing any form of
// arithmetic or transformation with currency in float form.
func NewAUDFromDollars(dollars float64) AUD {
	return AUD(math.Round(dollars * 100))
}

// String returns a human readable string representatin of the currency,
// like $12.34, $0.40, or $1500.00
func (a AUD) String() string {
	if a < 100 {
		return "$0." + centify(strconv.FormatInt(int64(a), 10))
	}

	return "$" + strconv.FormatInt(int64(a/100), 10) + "." +
		centify(strconv.FormatInt(int64(a%100), 10))
}

func centify(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}
