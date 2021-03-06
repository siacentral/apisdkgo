package scprime

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/siacentral/apisdkgo/scprime/types"
)

type (
	getConnectionResp struct {
		APIResponse
		Report types.ConnectionReport `json:"report"`
	}
)

//GetHostConnectivity checks that a host is running and connectable at the provided netaddress
func (a *APIClient) GetHostConnectivity(netaddress string) (report types.ConnectionReport, err error) {
	var resp getConnectionResp

	code, err := a.makeAPIRequest(http.MethodGet, fmt.Sprintf("/troubleshoot/%s", url.PathEscape(netaddress)), nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 {
		err = errors.New(resp.Message)
		return
	}

	report = resp.Report

	return
}
