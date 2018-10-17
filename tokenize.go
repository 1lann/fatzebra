package fatzebra

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// TokenizeCode represents a tokenize return code.
type TokenizeCode int

// UnmarshalJSON unmarshals the tokenize response code from JSON.
func (t *TokenizeCode) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*int)(t))
}

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
	h := hmac.New(md5.New, c.secret)
	h.Write([]byte(redirectURL))
	return hex.EncodeToString(h.Sum(nil))
}

// DirectTokenizeResult represents the result of a direct tokenize request.
type DirectTokenizeResult struct {
	TokenizeCode     TokenizeCode `json:"r"`
	CardToken        string       `json:"token"`
	VerificationCode string       `json:"v"`
	Errors           []string     `json:"errors[]"`
}

// Validate returns the validation result of the given verification code.
func (d *DirectTokenizeResult) Validate(c *Client) error {
	v, err := hex.DecodeString(d.VerificationCode)
	if err != nil {
		return err
	}

	data := []byte(strconv.Itoa(int(d.TokenizeCode)) + ":" + d.CardToken)

	h := hmac.New(md5.New, c.secret)
	h.Write(data)

	if !hmac.Equal(v, h.Sum(nil)) {
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
	var tokenResult DirectTokenizeResult

	if req.Method == http.MethodPost {
		if err := json.NewDecoder(req.Body).Decode(&tokenResult); err != nil {
			return nil, err
		}
	} else {
		query := req.URL.Query()
		retCode, err := strconv.Atoi(query.Get("r"))
		if err != nil {
			return nil, err
		}

		tokenResult = DirectTokenizeResult{
			TokenizeCode:     TokenizeCode(retCode),
			CardToken:        query.Get("token"),
			VerificationCode: query.Get("v"),
			Errors:           query["errors"],
		}
	}

	if err := tokenResult.Validate(c); err != nil {
		return nil, err
	}

	return &tokenResult, nil
}

// TokenizedCard represents a tokenized card.
type TokenizedCard struct {
	Token            string `json:"token"`
	CardHolder       string `json:"card_holder"`
	CardNumber       string `json:"card_number"`
	CardExpiry       string `json:"card_expiry"`
	CardType         string `json:"card_type"`
	CardCategory     string `json:"card_category"`
	CardSubcategory  string `json:"card_subcategory"`
	CardIssuer       string `json:"card_issuer"`
	CardCountry      string `json:"card_country"`
	Authorized       bool   `json:"authorized"`
	TransactionCount int    `json:"transaction_count"`
}

const (
	creditCardsV1 = "/v1.0/credit_cards/"
)

// tokenizedCardResult represents the result that is returned by a request
// for a tokenized card.
type tokenizedCardResult struct {
	Successful bool           `json:"successful"`
	Response   *TokenizedCard `json:"response"`
	Errors     []string       `json:"errors"`
}

// GetTokenizedCard returns a tokenized card given its token.
func (c *Client) GetTokenizedCard(ctx context.Context,
	token string) (*TokenizedCard, error) {
	req, err := http.NewRequest(http.MethodGet,
		"https://"+c.host+creditCardsV1+url.PathEscape(token), nil)
	if err != nil {
		panic(err)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("fatzebra: not OK: " + resp.Status)
	}

	var result tokenizedCardResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if !result.Successful {
		return nil, errors.New("fatzebra: error from server: " +
			result.Errors[0])
	}

	return result.Response, nil
}
