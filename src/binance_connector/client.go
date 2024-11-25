package binance_connector

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/binance/binance-connector-go/handlers"
	"github.com/bitly/go-simplejson"
)

// TimeInForceType define time in force type of order
type TimeInForceType string

// UserDataEventType define spot user data event type
type UserDataEventType string

// Client define API client
type Client struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	HTTPClient *http.Client
	Debug      bool
	Logger     *log.Logger
	TimeOffset int64
	do         doFunc
}

type doFunc func(req *http.Request) (*http.Response, error)

// Globals
const (
	timestampKey  = "timestamp"
	signatureKey  = "signature"
	recvWindowKey = "recvWindow"
)

func currentTimestamp() int64 {
	return FormatTimestamp(time.Now())
}

// FormatTimestamp formats a time into Unix timestamp in milliseconds, as requested by Binance.
func FormatTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func (c *Client) debug(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v...)
	}
}

// Create client function for initialising new Binance client
func NewClient(apiKey string, secretKey string, baseURL ...string) *Client {
	url := "https://api.binance.com"

	if len(baseURL) > 0 {
		url = baseURL[0]
	}

	return &Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    url,
		HTTPClient: http.DefaultClient,
		Logger:     log.New(os.Stderr, Name, log.LstdFlags),
	}
}

func (c *Client) parseRequest(r *request, opts ...RequestOption) (err error) {

	//如果不是api系列的request，要依request換掉endpoint
	if !strings.HasPrefix(r.endpoint, "/api/") {
		if strings.HasPrefix(r.endpoint, "/fapi/") {
			c.BaseURL = "https://fapi.binance.com"
		} else if strings.HasPrefix(r.endpoint, "/papi/") {
			c.BaseURL = "https://papi.binance.com"
		}
	}

	// set request options from user
	for _, opt := range opts {
		opt(r)
	}
	err = r.validate()
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, r.endpoint)
	if r.recvWindow > 0 {
		r.setParam(recvWindowKey, r.recvWindow)
	}
	if r.secType == secTypeSigned {
		r.setParam(timestampKey, currentTimestamp()-c.TimeOffset)
	}
	queryString := r.query.Encode()
	body := &bytes.Buffer{}
	bodyString := r.form.Encode()
	header := http.Header{}
	if r.header != nil {
		header = r.header.Clone()
	}
	header.Set("User-Agent", fmt.Sprintf("%s/%s", Name, Version))
	if bodyString != "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
		body = bytes.NewBufferString(bodyString)
	}
	if r.secType == secTypeAPIKey || r.secType == secTypeSigned {
		header.Set("X-MBX-APIKEY", c.APIKey)
	}

	if r.secType == secTypeSigned {
		raw := fmt.Sprintf("%s%s", queryString, bodyString)
		mac := hmac.New(sha256.New, []byte(c.SecretKey))
		_, err = mac.Write([]byte(raw))
		if err != nil {
			return err
		}
		v := url.Values{}
		v.Set(signatureKey, fmt.Sprintf("%x", (mac.Sum(nil))))
		if queryString == "" {
			queryString = v.Encode()
		} else {
			queryString = fmt.Sprintf("%s&%s", queryString, v.Encode())
		}
	}
	if queryString != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryString)
	}
	c.debug("full url: %s, body: %s", fullURL, bodyString)
	r.fullURL = fullURL
	r.header = header
	r.body = body
	return nil
}

func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
	err = c.parseRequest(r, opts...)
	if r.endpoint != "/api/v3/order/cancelReplace" {
		if err != nil {
			return []byte{}, err
		}
	}

	req, err := http.NewRequest(r.method, r.fullURL, r.body)
	if err != nil {
		return []byte{}, err
	}
	req = req.WithContext(ctx)
	req.Header = r.header
	c.debug("request: %#v", req)
	f := c.do
	if f == nil {
		f = c.HTTPClient.Do
	}
	res, err := f(req)
	if err != nil {
		return []byte{}, err
	}
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		cerr := res.Body.Close()
		// Only overwrite the retured error if the original error was nil and an
		// error occurred while closing the body.
		if err == nil && cerr != nil {
			err = cerr
		}
	}()
	c.debug("response: %#v", res)
	c.debug("response body: %s", string(data))
	c.debug("response status code: %d", res.StatusCode)

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := new(handlers.APIError)
		e := json.Unmarshal(data, apiErr)
		if e != nil {
			c.debug("failed to unmarshal json: %s", e)
		}
		if r.endpoint != "/api/v3/order/cancelReplace" {
			return nil, apiErr
		}
	}
	return data, nil
}

func newJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// Account Endpoints:
func (c *Client) NewTestNewOrder() *TestNewOrder {
	return &TestNewOrder{c: c}
}

func (c *Client) NewCreateOrderService() *CreateOrderService {
	return &CreateOrderService{c: c}
}

func (c *Client) NewCancelOrderService() *CancelOrderService {
	return &CancelOrderService{c: c}
}

func (c *Client) NewCancelOpenOrdersService() *CancelOpenOrdersService {
	return &CancelOpenOrdersService{c: c}
}

func (c *Client) NewGetOrderService() *GetOrderService {
	return &GetOrderService{c: c}
}

func (c *Client) NewCancelReplaceService() *CancelReplaceService {
	return &CancelReplaceService{c: c}
}

func (c *Client) NewGetOpenOrdersService() *GetOpenOrdersService {
	return &GetOpenOrdersService{c: c}
}

func (c *Client) NewGetAllOrdersService() *GetAllOrdersService {
	return &GetAllOrdersService{c: c}
}

func (c *Client) NewNewOCOService() *NewOCOService {
	return &NewOCOService{c: c}
}

func (c *Client) NewCancelOCOService() *CancelOCOService {
	return &CancelOCOService{c: c}
}

func (c *Client) NewQueryOCOService() *QueryOCOService {
	return &QueryOCOService{c: c}
}

func (c *Client) NewQueryAllOCOService() *QueryAllOCOService {
	return &QueryAllOCOService{c: c}
}

func (c *Client) NewQueryOpenOCOService() *QueryOpenOCOService {
	return &QueryOpenOCOService{c: c}
}

func (c *Client) NewGetAccountService() *GetAccountService {
	return &GetAccountService{c: c}
}

func (c *Client) NewGetMyTradesService() *GetMyTradesService {
	return &GetMyTradesService{c: c}
}

func (c *Client) NewGetQueryCurrentOrderCountUsageService() *GetQueryCurrentOrderCountUsageService {
	return &GetQueryCurrentOrderCountUsageService{c: c}
}

func (c *Client) NewGetQueryPreventedMatchesService() *GetQueryPreventedMatchesService {
	return &GetQueryPreventedMatchesService{c: c}
}

// Market Endpoints:
func (c *Client) NewPingService() *Ping {
	return &Ping{c: c}
}

func (c *Client) NewServerTimeService() *ServerTime {
	return &ServerTime{c: c}
}

func (c *Client) NewExchangeInfoService() *ExchangeInfo {
	return &ExchangeInfo{c: c}
}

func (c *Client) NewOrderBookService() *OrderBook {
	return &OrderBook{c: c}
}

func (c *Client) NewRecentTradesListService() *RecentTradesList {
	return &RecentTradesList{c: c}
}

func (c *Client) NewHistoricalTradeLookupService() *HistoricalTradeLookup {
	return &HistoricalTradeLookup{c: c}
}

func (c *Client) NewAggTradesListService() *AggTradesList {
	return &AggTradesList{c: c}
}

func (c *Client) NewKlinesService() *Klines {
	return &Klines{c: c}
}

func (c *Client) NewUIKlinesService() *UiKlines {
	return &UiKlines{c: c}
}

func (c *Client) NewAvgPriceService() *AvgPrice {
	return &AvgPrice{c: c}
}

func (c *Client) NewTicker24hrService() *Ticker24hr {
	return &Ticker24hr{c: c}
}

func (c *Client) NewTickerPriceService() *TickerPrice {
	return &TickerPrice{c: c}
}

func (c *Client) NewTickerBookTickerService() *TickerBookTicker {
	return &TickerBookTicker{c: c}
}

func (c *Client) NewTickerService() *Ticker {
	return &Ticker{c: c}
}

func (c *Client) NewGetAllMarginAssetsService() *GetAllMarginAssetsService {
	return &GetAllMarginAssetsService{c: c}
}

func (c *Client) NewGetAllMarginPairsService() *GetAllMarginPairsService {
	return &GetAllMarginPairsService{c: c}
}

func (c *Client) NewQueryMarginPriceIndexService() *QueryMarginPriceIndexService {
	return &QueryMarginPriceIndexService{c: c}
}

func (c *Client) NewMarginAccountNewOrderService() *MarginAccountNewOrderService {
	return &MarginAccountNewOrderService{c: c}
}

func (c *Client) NewMarginAccountCancelOrderService() *MarginAccountCancelOrderService {
	return &MarginAccountCancelOrderService{c: c}
}

func (c *Client) NewMarginAccountCancelAllOrdersService() *MarginAccountCancelAllOrdersService {
	return &MarginAccountCancelAllOrdersService{c: c}
}

func (c *Client) NewCrossMarginTransferHistoryService() *CrossMarginTransferHistoryService {
	return &CrossMarginTransferHistoryService{c: c}
}

func (c *Client) NewInterestHistoryService() *InterestHistoryService {
	return &InterestHistoryService{c: c}
}

func (c *Client) NewForceLiquidationRecordService() *ForceLiquidationRecordService {
	return &ForceLiquidationRecordService{c: c}
}

func (c *Client) NewCrossMarginAccountDetailService() *CrossMarginAccountDetailService {
	return &CrossMarginAccountDetailService{c: c}
}

func (c *Client) NewMarginAccountOrderService() *MarginAccountOrderService {
	return &MarginAccountOrderService{c: c}
}

func (c *Client) NewMarginAccountOpenOrderService() *MarginAccountOpenOrderService {
	return &MarginAccountOpenOrderService{c: c}
}

func (c *Client) NewMarginAccountAllOrderService() *MarginAccountAllOrderService {
	return &MarginAccountAllOrderService{c: c}
}

func (c *Client) NewMarginAccountNewOCOService() *MarginAccountNewOCOService {
	return &MarginAccountNewOCOService{c: c}
}

func (c *Client) NewMarginAccountCancelOCOService() *MarginAccountCancelOCOService {
	return &MarginAccountCancelOCOService{c: c}
}

func (c *Client) NewMarginAccountQueryOCOService() *MarginAccountQueryOCOService {
	return &MarginAccountQueryOCOService{c: c}
}

func (c *Client) NewMarginAccountQueryAllOCOService() *MarginAccountQueryAllOCOService {
	return &MarginAccountQueryAllOCOService{c: c}
}

func (c *Client) NewMarginAccountQueryOpenOCOService() *MarginAccountQueryOpenOCOService {
	return &MarginAccountQueryOpenOCOService{c: c}
}

func (c *Client) NewMarginAccountQueryTradeListService() *MarginAccountQueryTradeListService {
	return &MarginAccountQueryTradeListService{c: c}
}

func (c *Client) NewMarginAccountQueryMaxBorrowService() *MarginAccountQueryMaxBorrowService {
	return &MarginAccountQueryMaxBorrowService{c: c}
}

func (c *Client) NewMarginAccountQueryMaxTransferOutAmountService() *MarginAccountQueryMaxTransferOutAmountService {
	return &MarginAccountQueryMaxTransferOutAmountService{c: c}
}

func (c *Client) NewMarginAccountSummaryService() *MarginAccountSummaryService {
	return &MarginAccountSummaryService{c: c}
}

func (c *Client) NewMarginIsolatedAccountInfoService() *MarginIsolatedAccountInfoService {
	return &MarginIsolatedAccountInfoService{c: c}
}

func (c *Client) NewMarginIsolatedAccountDisableService() *MarginIsolatedAccountDisableService {
	return &MarginIsolatedAccountDisableService{c: c}
}

func (c *Client) NewMarginIsolatedAccountEnableService() *MarginIsolatedAccountEnableService {
	return &MarginIsolatedAccountEnableService{c: c}
}

func (c *Client) NewMarginIsolatedAccountLimitService() *MarginIsolatedAccountLimitService {
	return &MarginIsolatedAccountLimitService{c: c}
}

func (c *Client) NewAllIsolatedMarginSymbolService() *AllIsolatedMarginSymbolService {
	return &AllIsolatedMarginSymbolService{c: c}
}

func (c *Client) NewMarginToggleBnbBurnService() *MarginToggleBnbBurnService {
	return &MarginToggleBnbBurnService{c: c}
}

func (c *Client) NewMarginBnbBurnStatusService() *MarginBnbBurnStatusService {
	return &MarginBnbBurnStatusService{c: c}
}

func (c *Client) NewMarginInterestRateHistoryService() *MarginInterestRateHistoryService {
	return &MarginInterestRateHistoryService{c: c}
}

func (c *Client) NewMarginCrossMarginFeeService() *MarginCrossMarginFeeService {
	return &MarginCrossMarginFeeService{c: c}
}

func (c *Client) NewMarginIsolatedMarginFeeService() *MarginIsolatedMarginFeeService {
	return &MarginIsolatedMarginFeeService{c: c}
}

func (c *Client) NewMarginIsolatedMarginTierService() *MarginIsolatedMarginTierService {
	return &MarginIsolatedMarginTierService{c: c}
}

func (c *Client) NewMarginCurrentOrderCountService() *MarginCurrentOrderCountService {
	return &MarginCurrentOrderCountService{c: c}
}

func (c *Client) NewMarginCrossCollateralRatioService() *MarginCrossCollateralRatioService {
	return &MarginCrossCollateralRatioService{c: c}
}

func (c *Client) NewMarginSmallLiabilityExchangeCoinListService() *MarginSmallLiabilityExchangeCoinListService {
	return &MarginSmallLiabilityExchangeCoinListService{c: c}
}

func (c *Client) NewMarginSmallLiabilityExchangeService() *MarginSmallLiabilityExchangeService {
	return &MarginSmallLiabilityExchangeService{c: c}
}

func (c *Client) NewMarginSmallLiabilityExchangeHistoryService() *MarginSmallLiabilityExchangeHistoryService {
	return &MarginSmallLiabilityExchangeHistoryService{c: c}
}

// User Data Streams:
func (c *Client) NewCreateListenKeyService() *CreateListenKey {
	return &CreateListenKey{c: c}
}

func (c *Client) NewPingUserStream() *PingUserStream {
	return &PingUserStream{c: c}
}

func (c *Client) NewCloseUserStream() *CloseUserStream {
	return &CloseUserStream{c: c}
}

// 新增統一帳戶功能
func (c *Client) GetUMPositionService() *GetUMPositionService {
	return &GetUMPositionService{c: c}
}
func (c *Client) GetUMAccountAssetService() *GetUMAccountAssetService {
	return &GetUMAccountAssetService{c: c}
}
func (c *Client) GetUMAccountBalanceService() *GetUMAccountBalanceService {
	return &GetUMAccountBalanceService{c: c}
}

func (c *Client) GetUMPositionRiskService() *GetUMPositionRiskService {
	return &GetUMPositionRiskService{c: c}
}

func (c *Client) GetUMNewOrderService() *GetUMNewOrderService {
	return &GetUMNewOrderService{c: c}
}

func (c *Client) GetUMOrderService() *GetUMOrderService {
	return &GetUMOrderService{c: c}
}

func (c *Client) GetUMUserTradeService() *GetUMUserTradeService {
	return &GetUMUserTradeService{c: c}
}
