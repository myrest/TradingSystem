package bitunix_feature

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

type doFunc func(req *http.Request) (*http.Response, error)

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
	//url := "https://openapi.bitunix.com"
	url := "https://fapi.bitunix.com"

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

	//Todo: Post的還沒驗證過，先驗Get
	if r.method == http.MethodPost {
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
	}

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
		}
	}

	// Convert the map to a JSON string
	jsonBytes, err := json.Marshal(queryMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
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
			queryParamsStr += fmt.Sprintf("%s%s", k, v)
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

func SortJSONKeys(jsonStr string) (string, error) {
	//如果為空，就直接回傳
	if (jsonStr == "") || (jsonStr == "{}") {
		return jsonStr, nil
	}
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

// region 新增功能
//func (c *Client) GetLeverageService() *GetLeverageService {
//	return &GetLeverageService{c: c}
//}

// endregion
