package bitunix

import (
	"TradingSystem/src/common"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// API Endpoints
const (
	baseApiUrl = "https://openapi.bitunix.com"
)

// Side type of order
type SideType int //Side (1 Sell 2 Buy)
// Type of order
type OrderType int //Order Type(1:Limit 2:Market)

// PositionSide type of order
type PositionSideType string
type MarginTradingType string

type OrderStatus string

type OrderSpecType string

type OrderWorkingType string

const (
	SellSideType SideType = 1
	BuySideType  SideType = 2

	LimitOrderType  OrderType = 1
	MarketOrderType OrderType = 2

	//以下的還沒確認
	ShortPositionSideType PositionSideType = "SHORT"
	LongPositionSideType  PositionSideType = "LONG"
	BothPositionSideType  PositionSideType = "BOTH"

	MarginIsolated MarginTradingType = "ISOLATED"
	MarginCrossed  MarginTradingType = "CROSSED"

	NewOrderStatus             OrderStatus = "NEW"
	PartiallyFilledOrderStatus OrderStatus = "PARTIALLY_FILLED"
	FilledOrderStatus          OrderStatus = "FILLED"
	CanceledOrderStatus        OrderStatus = "CANCELED"
	ExpiredOrderStatus         OrderStatus = "EXPIRED"

	NewOrderSpecType        OrderSpecType = "NEW"
	CanceledOrderSpecType   OrderSpecType = "CANCELED"
	CalculatedOrderSpecType OrderSpecType = "CALCULATED"
	ExpiredOrderSpecType    OrderSpecType = "EXPIRED"
	TradeOrderSpecType      OrderSpecType = "TRADE"

	MarkOrderWorkingType     OrderWorkingType = "MARK_PRICE"
	ContractOrderWorkingType OrderWorkingType = "CONTRACT_PRICE"
	IndexOrderWorkingType    OrderWorkingType = "INDEX_PRICE"
)

type Interval string

const (
	Interval1  Interval = "1m"
	Interval3  Interval = "3m"
	Interval5  Interval = "5m"
	Interval15 Interval = "15m"
	Interval30 Interval = "30m"

	Interval60  Interval = "1h"
	Interval2h  Interval = "2h"
	Interval4h  Interval = "4h"
	Interval6h  Interval = "6h"
	Interval8h  Interval = "8h"
	Interval12h Interval = "12h"

	Interval1d Interval = "1d"
	Interval3d Interval = "3d"

	Interval1w Interval = "1w"

	Interval1M Interval = "1M"
)

func getApiEndpoint() string {
	return baseApiUrl
}

type doFunc func(*http.Request) (*http.Response, error)

type bitunixAPIRawResponse struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
}

func (c *bitunixAPIRawResponse) GetError() error {
	if !c.Success {
		return fmt.Errorf("<Bitunix API> code=%s, msg=%s", c.Code, c.Msg)
	} else {
		return nil
	}
}

// Client define API client
type Client struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	UserAgent  string
	HTTPClient *http.Client
	Debug      bool
	Logger     *log.Logger
	TimeOffset int64
	do         doFunc
}

// Init Api Client from apiKey & secretKey
func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    getApiEndpoint(),
		UserAgent:  "Bitunix/golang",
		HTTPClient: http.DefaultClient,
		Logger:     log.New(os.Stderr, "bitunix-golang", log.LstdFlags),
	}
}
func (c *Client) debug(message string, args ...interface{}) {
	if c.Debug {
		c.Logger.Printf(message, args...)
	}
}

// urlValuesToJSON converts url.Values to a JSON string.
func urlValuesToJSON(v url.Values) (string, error) {
	// Parse the query string
	values, err := url.ParseQuery(v.Encode())
	if err != nil {
		return "", err
	}

	queryMap := make(map[string]interface{})

	// Iterate over the parsed values and store them in the map
	for key, value := range values {
		// Use the first value if there are multiple values for a key
		if len(value) > 0 {
			if (key == "side") || (key == "type") { //side及type為數字
				queryMap[key], _ = strconv.Atoi(value[0])
			} else {
				queryMap[key] = value[0] // 或使用 value 直接作為 []string
			}
			queryMap[key] = value[0] // 或使用 value 直接作為 []string
		}
	}

	// Convert the map to a JSON string
	jsonBytes, err := json.Marshal(queryMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (c *Client) parseRequest(r *request, opts ...RequestOption) (err error) {
	for _, opt := range opts {
		opt(r)
	}

	err = r.validate()
	if err != nil {
		return err
	}

	if r.nonce == "" {
		r.nonce = common.GenerateRandomString(32)
	}

	jsonBody := ""
	jsonObject := url.Values{}
	if len(r.query) > 0 {
		jsonObject = r.query
	} else if len(r.form) > 0 {
		jsonObject = r.form
	}

	if len(jsonObject) > 0 {
		jsonBody, err = urlValuesToJSON(jsonObject)
		if err != nil {
			return err
		}
	}
	r.query.Encode()
	r.form = url.Values{}  // Clear the form
	r.query = url.Values{} // Clear the query

	timestamp := time.Now().UnixNano() / 1e6
	//sign要寫到Header裏面
	sign, err := MakeSign(c.APIKey, r.nonce, timestamp, r.query, jsonBody, c.SecretKey)
	if err != nil {
		return err
	}

	queryString := r.query.Encode()
	body := &bytes.Buffer{}
	//bodyString := r.form.Encode() //不使用form
	bodyString := jsonBody
	header := http.Header{}
	if r.header != nil {
		header = r.header.Clone()
	}
	header.Set("Content-Type", "application/json")
	header.Add("api-key", c.APIKey)
	header.Add("nonce", r.nonce)
	header.Add("timestamp", strconv.FormatInt(timestamp, 10))
	header.Add("sign", sign)

	if bodyString != "" {
		body = bytes.NewBufferString(bodyString)
	}

	fullUrl := fmt.Sprintf("%s%s", c.BaseURL, r.endpoint)

	if queryString != "" {
		fullUrl = fmt.Sprintf("%s?%s", fullUrl, queryString)
	}

	r.fullUrl = fullUrl
	r.header = header
	r.body = body
	return nil
}

func MakeSign(apiKey string, nonce string, timestamp int64, queryParams url.Values, jsonBody, secretKey string) (string, error) {
	// Step 1: Prepare queryParams string
	var queryParamsStr string
	// Get the keys and sort them in ascending order
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort keys in ascending order

	// Construct the queryParams string
	for _, k := range keys {
		for _, v := range queryParams[k] {
			queryParamsStr += fmt.Sprintf("%s=%s", k, v)
		}
	}

	// Step 2: Prepare body string
	jsonBody, err := SortJSONKeys(jsonBody)
	if err != nil {
		return "", err
	}

	bodyStr := strings.ReplaceAll(jsonBody, " ", "") //去空白

	// Step 3: Create the digest
	digest := sha256.Sum256([]byte(nonce + strconv.FormatInt(timestamp, 10) + apiKey + queryParamsStr + bodyStr))
	digestHex := hex.EncodeToString(digest[:])

	// Step 4: Create the signature
	signature := sha256.Sum256([]byte(digestHex + secretKey))
	sign := hex.EncodeToString(signature[:])

	return sign, nil
}

// SortJSONKeys takes a JSON string, sorts its keys, and returns a new JSON string with sorted keys.
func SortJSONKeys(jsonStr string) (string, error) {
	// Unmarshal the original JSON string into a map
	var originalMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &originalMap); err != nil {
		return "", err
	}

	// Create a sorted slice of keys
	var keys []string
	for key := range originalMap {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sort the keys

	// Create a new map to hold the sorted key-value pairs
	sortedMap := make(map[string]interface{})
	for _, key := range keys {
		sortedMap[key] = originalMap[key]
	}

	// Marshal the sorted map back to a JSON string
	sortedJSON, err := json.Marshal(sortedMap)
	if err != nil {
		return "", err
	}

	return string(sortedJSON), nil
}

func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
	err = c.parseRequest(r, opts...)
	if err != nil {
		return []byte{}, err
	}
	req, err := http.NewRequest(r.method, r.fullUrl, r.body)
	if err != nil {
		return []byte{}, err
	}
	req = req.WithContext(ctx)
	req.Header = r.header
	c.debug("request url: %#v", req.URL.String())
	c.debug("body: %s", r.body)
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
		if err == nil && cerr != nil {
			err = cerr
		}
	}()
	c.debug("response body: %s", string(data))
	c.debug("response status code: %d", res.StatusCode)

	apiErr := new(bitunixAPIRawResponse)
	json.Unmarshal(data, apiErr)

	err = apiErr.GetError()

	return data, err
}

type GetServerTimeService struct {
	c *Client
}

func (s *GetServerTimeService) Do(ctx context.Context, opts ...RequestOption) (res int64, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/server/time"}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return 0, err
	}

	resp := new(struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data map[string]int64 `json:"data"`
	})

	err = json.Unmarshal(data, &resp)

	if err != nil {
		return 0, err
	}

	res = resp.Data["serverTime"]

	return res, nil
}

func (c *Client) NewGetOpenPositionsService() *GetOpenPositionsService {
	return &GetOpenPositionsService{c: c}
}

func (c *Client) NewGetBalanceService() *GetBalanceService {
	return &GetBalanceService{c: c}
}

func (c *Client) NewGetTradingPairService() *GetTradingPairService {
	return &GetTradingPairService{c: c}
}

// 以下未使用
func (c *Client) NewGetServerTimeService() *GetServerTimeService {
	return &GetServerTimeService{c: c}
}

func (c *Client) NewCreateOrderService() *CreateOrderService {
	return &CreateOrderService{c: c}
}

func (c *Client) NewCancelOrderService() *CancelOrderService {
	return &CancelOrderService{c: c}
}

func (c *Client) NewGetOrderService() *GetOrderService {
	return &GetOrderService{c: c}
}

func (c *Client) NewGetOpenOrdersService() *GetOpenOrdersService {
	return &GetOpenOrdersService{c: c}
}
