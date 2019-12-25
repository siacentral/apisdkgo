package apiclient

import (
	"errors"
	"fmt"

	"github.com/siacentral/apiclient/types"

	siatypes "gitlab.com/NebulousLabs/Sia/types"
)

type (
	getAddressesResp struct {
		APIResponse
		Addresses []types.AddressUsage `json:"addresses"`
	}

	getFeesResp struct {
		APIResponse
		Minimum    siatypes.Currency `json:"minimum"`
		Maximum    siatypes.Currency `json:"maximum"`
		SiaCentral siatypes.Currency `json:"sia_central"`
	}

	//GetTransactionsResp holds balance and transactions for an address or set of addresses
	GetTransactionsResp struct {
		APIResponse
		Unspent                 siatypes.Currency         `json:"unspent_total"`
		UnspentOutputs          []types.SiacoinOutput     `json:"unspent_outputs"`
		Transactions            []types.WalletTransaction `json:"transactions"`
		UnconfirmedTransactions []types.WalletTransaction `json:"unconfirmed_transactions"`
	}
)

//GetTransactionFees gets the current transaction fees of the Sia network
func GetTransactionFees() (min, max, internal siatypes.Currency, err error) {
	var resp getFeesResp

	code, err := makeAPIRequest(HTTPGet, "/wallet/fees", nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	min = resp.Minimum
	max = resp.Maximum
	internal = resp.SiaCentral

	return
}

//FindAddressBalance gets all unspent outputs and the last n transactions for a list of addresses
func FindAddressBalance(limit, page int, addresses []string) (resp GetTransactionsResp, err error) {
	if len(addresses) > 10000 {
		err = errors.New("maximum of 10000 addresses")
		return
	}

	code, err := makeAPIRequest(HTTPPost, fmt.Sprintf("/wallet/addresses?limit=%d&page=%d", limit, page), map[string]interface{}{
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

//FindUsedAddresses gets all addresses that have been seen in a transaction on the blockchain
func FindUsedAddresses(addresses []string) (used []types.AddressUsage, err error) {
	var resp getAddressesResp

	if len(addresses) > 10000 {
		err = errors.New("maximum of 10000 addresses")
		return
	}

	code, err := makeAPIRequest(HTTPPost, "/wallet/addresses/used", map[string]interface{}{
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

//GetAddressBalance gets all unspent outputs and the last n transactions of an address
func GetAddressBalance(limit, page int, address string) (resp GetTransactionsResp, err error) {
	code, err := makeAPIRequest(HTTPGet, fmt.Sprintf("/wallet/addresses/%s", address), nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	return
}

//Method:  "POST",
//Pattern: "/broadcast",
