package sia

import (
	"errors"
	"fmt"
	"net/http"

	"go.sia.tech/siad/types"
)

type (
	getAddressesResp struct {
		APIResponse
		Addresses []AddressUsage `json:"addresses"`
	}

	apiFees struct {
		Address string         `json:"address"`
		Fee     types.Currency `json:"fee"`
	}

	getFeesResp struct {
		APIResponse
		Minimum types.Currency `json:"minimum"`
		Maximum types.Currency `json:"maximum"`
		API     apiFees        `json:"api"`
	}

	//GetTransactionsResp holds balance and transactions for an address or set of addresses
	GetTransactionsResp struct {
		APIResponse
		UnspentSiacoins         types.Currency  `json:"unspent_siacoins"`
		UnspentSiafunds         types.Currency  `json:"unspent_siafunds"`
		UnspentSiacoinOutputs   []SiacoinOutput `json:"unspent_siacoin_outputs"`
		UnspentSiafundOutputs   []SiafundOutput `json:"unspent_siafund_outputs"`
		Transactions            []Transaction   `json:"transactions"`
		UnconfirmedTransactions []Transaction   `json:"unconfirmed_transactions"`
		SiafundClaim            types.Currency  `json:"siafund_claim"`
	}
)

// GetTransactionFees gets the current transaction fees of the Sia network
func (a *APIClient) GetTransactionFees() (min, max types.Currency, err error) {
	var resp getFeesResp

	code, err := a.makeAPIRequest(http.MethodGet, "/wallet/fees", nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	min = resp.Minimum
	max = resp.Maximum

	return
}

// GetAPIFees gets the current transaction fee and payout address of the Sia Central API
func (a *APIClient) GetAPIFees() (fee types.Currency, address string, err error) {
	var resp getFeesResp

	code, err := a.makeAPIRequest(http.MethodGet, "/wallet/fees", nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	fee = resp.API.Fee
	address = resp.API.Address

	return
}

// FindAddressBalance gets all unspent outputs and the last n transactions for a list of addresses
func (a *APIClient) FindAddressBalance(limit, page int, addresses []string) (resp GetTransactionsResp, err error) {
	if len(addresses) > 10000 {
		err = errors.New("maximum of 10000 addresses")
		return
	}

	code, err := a.makeAPIRequest(http.MethodPost, fmt.Sprintf("/wallet/addresses?limit=%d&page=%d", limit, page), map[string]interface{}{
		"addresses": addresses,
	}, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	return
}

// FindUsedAddresses gets all addresses that have been seen in a transaction on the blockchain
func (a *APIClient) FindUsedAddresses(addresses []string) (used []AddressUsage, err error) {
	var resp getAddressesResp

	if len(addresses) > 10000 {
		err = errors.New("maximum of 10000 addresses")
		return
	}

	code, err := a.makeAPIRequest(http.MethodPost, "/wallet/addresses/used", map[string]interface{}{
		"addresses": addresses,
	}, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	used = resp.Addresses

	return
}

// GetAddressBalance gets all unspent outputs and the last n transactions of an address
func (a *APIClient) GetAddressBalance(limit, page int, address string) (resp GetTransactionsResp, err error) {
	code, err := a.makeAPIRequest(http.MethodGet, fmt.Sprintf("/wallet/addresses/%s", address), nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	return
}

// BroadcastTransactionSet broadcasts the transaction set to the network
func (a *APIClient) BroadcastTransactionSet(transactions []types.Transaction) (err error) {
	var resp APIResponse

	code, err := a.makeAPIRequest(http.MethodPost, "/wallet/broadcast", map[string]interface{}{
		"transactions": transactions,
	}, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	return
}
