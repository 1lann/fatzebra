package fatzebra

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// purchasesListingResult represents a result returned by listing or
// retrieving purchases.
type purchasesListingResult struct {
	Successful   bool        `json:"successful"`
	Response     []*Purchase `json:"response"`
	Errors       []string    `json:"errors"`
	Records      int         `json:"records"`
	TotalRecords int         `json:"total_records"`
	Page         int         `json:"page"`
	TotalPages   int         `json:"total_pages"`
}

type purchasesResult struct {
	Successful bool      `json:"successful"`
	Response   *Purchase `json:"response"`
	Errors     []string  `json:"errors"`
}

// GetPurchaseByReference retrieves a purchase by its reference code.
func (c *Client) GetPurchaseByReference(ctx context.Context,
	ref string) (*Purchase, error) {
	u := "https://" + c.host + purchasesV1 + "?reference=" +
		url.QueryEscape(ref)
	req, err := http.NewRequest(http.MethodGet, u, nil)
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

	var result purchasesResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("fatzebra: " + resp.Status)
		}

		return nil, err
	}

	if !result.Successful {
		return nil, errors.New("fatzebra: error from server: " +
			result.Errors[0])
	}

	return result.Response, nil
}
