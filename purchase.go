package fatzebra

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	purchasesV1 = "/v1.0/purchases"
)

// ValidationError represents a purchase validation error.
type ValidationError struct {
	Reference  string
	Errors     []string
	StatusCode int
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return "fatzebra: error response: " + e.Errors[0]
}

// Zap logs the information in the error response to a zap logger.
func (e *ValidationError) Zap(l *zap.SugaredLogger) {
	l.Named("fatzebra").Errorw("received unsuccessful result",
		"reference", e.Reference,
		"errors", e.Errors,
		"statusCode", e.StatusCode,
	)
}

// Possible errors returned by DoPurchase.
var (
	ErrExceedsMaximum = errors.New("fatzebra: transaction exceeds " +
		"maximum allowable")
)

// GenerateReference generates a random reference code, with a very very very
// small chance of collision. A prefix is recommended for customer support
// reasons. For instance, a person mistakes your company for another,
// they contact the wrong company. It makes it easier to tell that they're
// contacting the wrong company if the reference code is missing the prefix).
func GenerateReference(prefix string) string {
	min := new(big.Int)
	min.SetString("100000000000", 36)
	max := new(big.Int)
	max.SetString("zzzzzzzzzzzz", 36)

	result, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		panic(err)
	}
	result.Add(result, min)

	return prefix + strings.ToUpper(result.Text(36))
}

// PurchaseRequest represents a tokenized card purchase.
type PurchaseRequest struct {
	CardToken  string `json:"card_token,omitempty"`
	CVV        string `json:"cvv,omitempty"`
	Amount     AUD    `json:"amount"`
	Reference  string `json:"reference"`
	Capture    bool   `json:"capture"`
	CustomerIP string `json:"customer_ip"` // Yes, this is required.
}

// Purchase represents a purchase.
type Purchase struct {
	Authorization   string    `json:"authorization"`
	ID              string    `json:"id"`
	CardNumber      string    `json:"card_number"`
	CardHolder      string    `json:"card_holder"`
	CardExpiry      string    `json:"card_expiry"`
	CardToken       string    `json:"card_token"`
	CardType        string    `json:"card_type"`
	CardCategory    string    `json:"card_category"`
	CardSubcategory string    `json:"card_subcategory"`
	Amount          AUD       `json:"amount"`
	Successful      bool      `json:"successful"`
	Message         string    `json:"message"`
	Reference       string    `json:"reference"`
	Currency        string    `json:"currency"`
	TransactionID   string    `json:"transaction_id"`
	SettlementDate  string    `json:"settlement_date"`
	TransactionDate time.Time `json:"transaction_date"`
	ResponseCode    string    `json:"response_code"`
	Captured        bool      `json:"captured"`
	CapturedAmount  int       `json:"captured_amount"`
	RRN             string    `json:"rrn"`
	CVVMatch        string    `json:"cvv_match"`
}

// purchaseResult represents the result that is returned by a purchase.
type purchaseResult struct {
	Successful bool      `json:"successful"`
	Response   *Purchase `json:"response"`
	Errors     []string  `json:"errors"`
}

// DoPurchase issues a purchase with a tokenized card. Note that
// no error is returned if the authorization is rejected, check the
// response instead. Capture must be set to true for a purchase rather than a
// hold. Note that purchases with validation errors are not recorded in the
// transaction history.
func (c *Client) DoPurchase(ctx context.Context,
	purchase *PurchaseRequest) (*Purchase, error) {
	buf := new(bytes.Buffer)

	if purchase.Amount > c.maxAmount {
		return nil, ErrExceedsMaximum
	}

	if err := json.NewEncoder(buf).Encode(purchase); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost,
		"https://"+c.host+purchasesV1, buf)
	if err != nil {
		panic(err)
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("fatzebra: not OK: " + resp.Status)
	}

	var result purchaseResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("fatzebra: " + resp.Status)
		}

		return nil, err
	}

	if !result.Successful {
		return nil, &ValidationError{
			Errors:     result.Errors,
			StatusCode: resp.StatusCode,
		}
	}

	return result.Response, nil
}
