package sia

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.sia.tech/siad/types"
)

const (
	acceptContractsParam      = "acceptcontracts"
	onlineParam               = "online"
	benchmarkedParam          = "benchmarked"
	minAgeParam               = "minage"
	minUptimeParam            = "minuptime"
	minDurationParam          = "minduration"
	minStorageParam           = "minstorage"
	minUploadSpeedParam       = "minuploadspeed"
	minDownloadSpeedParam     = "mindownloadspeed"
	maxStoragePriceParam      = "maxstorageprice"
	maxUploadPriceParam       = "maxuploadprice"
	maxDownloadPriceParam     = "maxdownloadprice"
	maxContractPriceParam     = "maxcontractprice"
	maxBaseRPCPriceParam      = "maxbaserpcprice"
	maxSectorAccessPriceParam = "maxsectoraccessprice"

	sortParam = "sort"
	dirParam  = "dir"
)

// UniqueID is a unique identifier
type UniqueID types.Specifier

// MarshalJSON marshals an id as a hex string.
func (uid UniqueID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uid.String())
}

// String prints the uid in hex.
func (uid UniqueID) String() string {
	return hex.EncodeToString(uid[:])
}

type (
	// RPCPriceTable contains the cost of executing a RPC on a host. Each host can
	// set its own prices for the individual MDM instructions and RPC costs.
	RPCPriceTable struct {
		// UID is a specifier that uniquely identifies this price table
		UID UniqueID `json:"uid"`

		// Validity is a duration that specifies how long the host guarantees these
		// prices for and are thus considered valid.
		Validity time.Duration `json:"validity"`

		// HostBlockHeight is the block height of the host. This allows the renter
		// to create valid withdrawal messages in case it is not synced yet.
		HostBlockHeight types.BlockHeight `json:"hostblockheight"`

		// UpdatePriceTableCost refers to the cost of fetching a new price table
		// from the host.
		UpdatePriceTableCost types.Currency `json:"updatepricetablecost"`

		// AccountBalanceCost refers to the cost of fetching the balance of an
		// ephemeral account.
		AccountBalanceCost types.Currency `json:"accountbalancecost"`

		// FundAccountCost refers to the cost of funding an ephemeral account on the
		// host.
		FundAccountCost types.Currency `json:"fundaccountcost"`

		// LatestRevisionCost refers to the cost of asking the host for the latest
		// revision of a contract.
		// TODO: should this be free?
		LatestRevisionCost types.Currency `json:"latestrevisioncost"`

		// SubscriptionMemoryCost is the cost of storing a byte of data for
		// SubscriptionPeriod time.
		SubscriptionMemoryCost types.Currency `json:"subscriptionmemorycost"`

		// SubscriptionNotificationCost is the cost of a single notification on top
		// of what is charged for bandwidth.
		SubscriptionNotificationCost types.Currency `json:"subscriptionnotificationcost"`

		// MDM related costs
		//
		// InitBaseCost is the amount of cost that is incurred when an MDM program
		// starts to run. This doesn't include the memory used by the program data.
		// The total cost to initialize a program is calculated as
		// InitCost = InitBaseCost + MemoryTimeCost * Time
		InitBaseCost types.Currency `json:"initbasecost"`

		// MemoryTimeCost is the amount of cost per byte per time that is incurred
		// by the memory consumption of the program.
		MemoryTimeCost types.Currency `json:"memorytimecost"`

		// Cost values specific to the bandwidth consumption.
		DownloadBandwidthCost types.Currency `json:"downloadbandwidthcost"`
		UploadBandwidthCost   types.Currency `json:"uploadbandwidthcost"`

		// Cost values specific to the DropSectors instruction.
		DropSectorsBaseCost types.Currency `json:"dropsectorsbasecost"`
		DropSectorsUnitCost types.Currency `json:"dropsectorsunitcost"`

		// Cost values specific to the HasSector command.
		HasSectorBaseCost types.Currency `json:"hassectorbasecost"`

		// Cost values specific to the Read instruction.
		ReadBaseCost   types.Currency `json:"readbasecost"`
		ReadLengthCost types.Currency `json:"readlengthcost"`

		// Cost values specific to the RenewContract instruction.
		RenewContractCost types.Currency `json:"renewcontractcost"`

		// Cost values specific to the Revision command.
		RevisionBaseCost types.Currency `json:"revisionbasecost"`

		// SwapSectorCost is the cost of swapping 2 full sectors by root.
		SwapSectorCost types.Currency `json:"swapsectorcost"`

		// Cost values specific to the Write instruction.
		WriteBaseCost   types.Currency `json:"writebasecost"`   // per write
		WriteLengthCost types.Currency `json:"writelengthcost"` // per byte written
		WriteStoreCost  types.Currency `json:"writestorecost"`  // per byte / block of additional storage

		// TxnFee estimations.
		TxnFeeMinRecommended types.Currency `json:"txnfeeminrecommended"`
		TxnFeeMaxRecommended types.Currency `json:"txnfeemaxrecommended"`

		// ContractPrice is the additional fee a host charges when forming/renewing
		// a contract to cover the miner fees when submitting the contract and
		// revision to the blockchain.
		ContractPrice types.Currency `json:"contractprice"`

		// CollateralCost is the amount of money per byte the host is promising to
		// lock away as collateral when adding new data to a contract. It's paid out
		// to the host regardless of the outcome of the storage proof.
		CollateralCost types.Currency `json:"collateralcost"`

		// MaxCollateral is the maximum amount of collateral the host is willing to
		// put into a single file contract.
		MaxCollateral types.Currency `json:"maxcollateral"`

		// MaxDuration is the max duration for which the host is willing to form a
		// contract.
		MaxDuration types.BlockHeight `json:"maxduration"`

		// WindowSize is the minimum time in blocks the host requests the
		// renewWindow of a new contract to be.
		WindowSize types.BlockHeight `json:"windowsize"`

		// Registry related fields.
		RegistryEntriesLeft  uint64 `json:"registryentriesleft"`
		RegistryEntriesTotal uint64 `json:"registryentriestotal"`
	}

	getHostsResp struct {
		APIResponse
		Hosts []HostDetails `json:"hosts"`
	}

	getHostDetailResp struct {
		APIResponse
		Host HostDetails `json:"host"`
	}

	getAveragesResp struct {
		APIResponse
		Settings       HostConfig       `json:"settings"`
		PriceTable     RPCPriceTable    `json:"price_table"`
		Benchmarks     AvgHostBenchmark `json:"benchmarks"`
		BenchmarksRHP2 AvgHostBenchmark `json:"benchmarks_rhp2"`
	}

	HostFilter func(url.Values)
	HostSort   string
)

var (
	HostSortDateCreated        HostSort = "date_created"
	HostSortNetAddress         HostSort = "net_address"
	HostSortPublicKey          HostSort = "public_key"
	HostSortAcceptingContracts HostSort = "accepting_contracts"
	HostSortUptime             HostSort = "uptime"
	HostSortUploadSpeed        HostSort = "upload_speed"
	HostSortDownloadSpeed      HostSort = "download_speed"
	HostSortRemainingStorage   HostSort = "remaining_storage"
	HostSortTotalStorage       HostSort = "total_storage"
	HostSortUsedStorage        HostSort = "used_storage"
	HostSortAge                HostSort = "age"
	HostSortUtilization        HostSort = "utilization"
	HostSortContractPrice      HostSort = "contract_price"
	HostSortStoragePrice       HostSort = "storage_price"
	HostSortDownloadPrice      HostSort = "download_price"
	HostSortUploadPrice        HostSort = "upload_price"
)

// HostFilterAcceptingContracts sets the accepting contracts parameter for the host's query
func HostFilterAcceptingContracts(accepting bool) HostFilter {
	return func(v url.Values) {
		v.Set(acceptContractsParam, strconv.FormatBool(accepting))
	}
}

// HostFilterOnline sets the online parameter for the host's query
func HostFilterOnline(online bool) HostFilter {
	return func(v url.Values) {
		v.Set(onlineParam, strconv.FormatBool(online))
	}
}

// HostFilterBenchmarked sets the benchmark parameter for the host's query
func HostFilterBenchmarked(benchmarked bool) HostFilter {
	return func(v url.Values) {
		v.Set(benchmarkedParam, strconv.FormatBool(benchmarked))
	}
}

// HostFilterMinAge sets the min age for the host query
func HostFilterMinAge(age uint64) HostFilter {
	return func(v url.Values) {
		v.Set(minAgeParam, strconv.FormatUint(age, 10))
	}
}

// HostFilterMinUptime sets the min uptime for the host query
func HostFilterMinUptime(minUptime float64) HostFilter {
	return func(v url.Values) {
		v.Set(minUptimeParam, strconv.FormatFloat(minUptime, 'f', -1, 64))
	}
}

// HostFilterMinDuration sets the min contract duration for the host query
func HostFilterMinDuration(minDuration uint64) HostFilter {
	return func(v url.Values) {
		v.Set(minDurationParam, strconv.FormatUint(minDuration, 10))
	}
}

// HostFilterMinStorage sets the min available storage for the host query
func HostFilterMinStorage(minStorage uint64) HostFilter {
	return func(v url.Values) {
		v.Set(minStorageParam, strconv.FormatUint(minStorage, 10))
	}
}

// HostFilterMinUploadSpeed sets the min upload speed for the host query
func HostFilterMinUploadSpeed(minUploadSpeed uint64) HostFilter {
	return func(v url.Values) {
		v.Set(minUploadSpeedParam, strconv.FormatUint(minUploadSpeed, 10))
	}
}

// HostFilterMinDownloadSpeed sets the min download speed for the host query
func HostFilterMinDownloadSpeed(minDownloadSpeed uint64) HostFilter {
	return func(v url.Values) {
		v.Set(minDownloadSpeedParam, strconv.FormatUint(minDownloadSpeed, 10))
	}
}

// HostFilterMaxStoragePrice sets the max storage price for the host query
func HostFilterMaxStoragePrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxStoragePriceParam, price.String())
	}
}

// HostFilterMaxUploadPrice sets the max upload price for the host query
func HostFilterMaxUploadPrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxUploadPriceParam, price.String())
	}
}

// HostFilterMaxDownloadPrice sets the max download price for the host query
func HostFilterMaxDownloadPrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxDownloadPriceParam, price.String())
	}
}

// HostFilterMaxContractPrice sets the max contract price for the host query
func HostFilterMaxContractPrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxContractPriceParam, price.String())
	}
}

// HostFilterMaxBaseRPCPrice sets the max base RPC price for the host query
func HostFilterMaxBaseRPCPrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxBaseRPCPriceParam, price.String())
	}
}

// HostFilterSectorAccessPrice sets the sector access price for the host query
func HostFilterSectorAccessPrice(price types.Currency) HostFilter {
	return func(v url.Values) {
		v.Set(maxSectorAccessPriceParam, price.String())
	}
}

// HostFilterSort sets the sort order for the host's query
func HostFilterSort(field HostSort, desc bool) HostFilter {
	return func(v url.Values) {
		v.Set(sortParam, string(field))
		if desc {
			v.Set(dirParam, "desc")
		} else {
			v.Set(dirParam, "asc")
		}
	}
}

// GetNetworkAverages gets the average settings and benchmarks of all active hosts on the network
func (a *APIClient) GetNetworkAverages() (settings HostConfig, rhp3Bench AvgHostBenchmark, rhp2Bench AvgHostBenchmark, err error) {
	var resp getAveragesResp

	code, err := a.makeAPIRequest(http.MethodGet, "/hosts/network/averages", nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	settings = resp.Settings
	rhp3Bench = resp.Benchmarks
	rhp2Bench = resp.BenchmarksRHP2

	return
}

// GetActiveHosts gets all Sia hosts that have been successfully scanned in the last 24 hours
func (a *APIClient) GetActiveHosts(page, limit int, filters ...HostFilter) (hosts []HostDetails, err error) {
	var resp getHostsResp

	if page < 0 {
		page = 0
	}

	if limit < 0 || limit > 500 {
		limit = 500
	}

	values := make(url.Values)
	for _, v := range filters {
		v(values)
	}

	values.Add("page", strconv.Itoa(page))
	values.Add("limit", strconv.Itoa(limit))

	endpoint, _ := url.Parse("https://api.siacentral.com/v2/hosts")
	endpoint.RawQuery = values.Encode()

	code, err := a.makeAPIRequest(http.MethodGet, endpoint.String(), nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	hosts = resp.Hosts

	return
}

// GetHost finds a host matching the public key or netaddress
func (a *APIClient) GetHost(id string) (host HostDetails, err error) {
	var resp getHostDetailResp

	code, err := a.makeAPIRequest(http.MethodGet, fmt.Sprintf("/hosts/%s", url.PathEscape(id)), nil, &resp)

	if err != nil {
		return
	}

	if code < 200 || code >= 300 || resp.Type != "success" {
		err = errors.New(resp.Message)
		return
	}

	host = resp.Host

	return
}
