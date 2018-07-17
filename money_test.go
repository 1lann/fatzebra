package fatzebra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoney(t *testing.T) {
	dollars := NewAUDFromCents(1234)
	assert.Equal(t, "$12.34", dollars.String())
	assert.Equal(t, int64(1234), dollars.Cents())
	assert.Equal(t, dollars, NewAUDFromDollars(12.34))
	assert.Equal(t, dollars, NewAUDFromDollars(12.341))
	assert.Equal(t, dollars, NewAUDFromDollars(12.339))

	cleanDollars := NewAUDFromDollars(12.00)
	assert.Equal(t, "$12.00", cleanDollars.String())
	assert.Equal(t, int64(1200), cleanDollars.Cents())

	cleanZero := NewAUDFromCents(0)
	assert.Equal(t, "$0.00", cleanZero.String())
	assert.Equal(t, int64(0), cleanZero.Cents())

	smallA := NewAUDFromDollars(0.01)
	assert.Equal(t, "$0.01", smallA.String())
	assert.Equal(t, int64(1), smallA.Cents())

	smallB := NewAUDFromDollars(0.02)
	assert.Equal(t, "$0.02", smallB.String())
	assert.Equal(t, int64(2), smallB.Cents())

	smallC := smallA + smallB
	assert.Equal(t, "$0.03", smallC.String())
	assert.Equal(t, int64(3), smallC.Cents())

	dollars += smallC
	assert.Equal(t, "$12.37", dollars.String())
	assert.Equal(t, int64(1237), dollars.Cents())
}
