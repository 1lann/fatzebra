package fatzebra

import (
	"crypto/hmac"
	"encoding/hex"
	"net/http"
	"strconv"
)

// TokenizeCode represents a tokenize return code.
type TokenizeCode int

// Valid tokenize return codes.
const (
	TokenizeSuccessful          TokenizeCode = 1
	TokenizeValidationError     TokenizeCode = 97
	TokenizeInvalidVerification TokenizeCode = 99
	TokenizeGatewayError        TokenizeCode = 999
)

// String returns a string representation of a tokenize return code.
func (c TokenizeCode) String() string {
	switch c {
	case TokenizeSuccessful:
		return "Successful"
	case TokenizeValidationError:
		return "ValidationError"
	case TokenizeInvalidVerification:
		return "InvalidVerification"
	case TokenizeGatewayError:
		return "GatewayError"
	default:
		return "Unknown"
	}
}

// GetTokenizeHash returns the hash that is to be sent by the user
// to redirect them to the given URL.
func (c *Client) GetTokenizeHash(redirectURL string) string {
	return hex.EncodeToString(c.hasher.Sum([]byte(redirectURL)))
}

// DirectTokenizeResult represents the result of a direct tokenize request.
type DirectTokenizeResult struct {
	TokenizeCode     TokenizeCode
	CardToken        string
	VerificationCode string
}

// Validate returns the validation result of the given verification code.
func (d *DirectTokenizeResult) Validate(c *Client) error {
	v, err := hex.DecodeString(d.VerificationCode)
	if err != nil {
		return err
	}

	data := []byte(strconv.Itoa(int(d.TokenizeCode)) + ":" + d.CardToken)
	if !hmac.Equal(v, c.hasher.Sum(data)) {
		return ErrBadHash
	}

	return nil
}

// ParseDirectTokenizeResult parses and validates the direct tokenize result
// received as a request. Details such as the card number, card holder and
// expiry date are not provided as they're not validated, and thus should
// not be trusted.
func (c *Client) ParseDirectTokenizeResult(
	req *http.Request) (*DirectTokenizeResult, error) {
	query := req.URL.Query()
	retCode, err := strconv.Atoi(query.Get("r"))
	if err != nil {
		return nil, err
	}

	tokenResult := &DirectTokenizeResult{
		TokenizeCode:     TokenizeCode(retCode),
		CardToken:        query.Get("token"),
		VerificationCode: query.Get("v"),
	}
	if err := tokenResult.Validate(c); err != nil {
		return nil, err
	}

	return tokenResult, nil
}
