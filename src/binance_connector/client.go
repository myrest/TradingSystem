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
		} else if strings.HasPrefix(r.endpoint, "/sapi/") {
			c.BaseURL = "https://sapi.binance.com"
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

// region 新增功能
func (c *Client) GetLeverageService() *GetLeverageService {
	return &GetLeverageService{c: c}
}

func (c *Client) GetPositionService() *GetPositionService {
	return &GetPositionService{c: c}
}

func (c *Client) GetMarginTypeService() *GetMarginTypeService {
	return &GetMarginTypeService{c: c}
}

func (c *Client) GetAccountBalanceService() *GetAccountBalanceService {
	return &GetAccountBalanceService{c: c}
}

func (c *Client) GetUMPositionRiskService() *GetUMPositionRiskService {
	return &GetUMPositionRiskService{c: c}
}

func (c *Client) GetUMNewOrderService() *GetUMNewOrderService {
	return &GetUMNewOrderService{c: c}
}

func (c *Client) GetUMUserTradeService() *GetUMUserTradeService {
	return &GetUMUserTradeService{c: c}
}

// endregion
