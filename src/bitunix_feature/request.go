package bitunix_feature

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// API request
type request struct {
	method   string
	endpoint string
	query    url.Values
	form     url.Values
	nonce    string
	header   http.Header
	body     io.Reader
	fullUrl  string
}

// addParam add param with key/value to query string
func (r *request) addParam(key string, value interface{}) *request {
	if r.query == nil {
		r.query = url.Values{}
	}
	var strValue string
	switch v := value.(type) {
	case float64:
		// 使用 strconv.FormatFloat 避免科学记号
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		// 如果是 float32，同样使用 strconv.FormatFloat
		strValue = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		strValue = fmt.Sprintf("%v", value)
	}

	r.query.Add(key, strValue)
	return r
}

// addParam add param with key/value to query string
func (r *request) setParam(key string, value interface{}) *request {
	if r.query == nil {
		r.query = url.Values{}
	}
	var strValue string
	switch v := value.(type) {
	case float64:
		// 使用 strconv.FormatFloat 避免科学记号
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		// 如果是 float32，同样使用 strconv.FormatFloat
		strValue = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		strValue = fmt.Sprintf("%v", value)
	}

	r.query.Set(key, strValue)
	return r
}

func (r *request) SetFormParam(key string, value interface{}) *request { //Not used
	if r.form == nil {
		r.form = url.Values{}
	}

	if reflect.TypeOf(value).Kind() == reflect.Slice {
		v, err := json.Marshal(value)
		if err == nil {
			value = string(v)
		}
	}

	r.form.Set(key, fmt.Sprintf("%v", value))
	return r
}

func (r *request) validate() (err error) {
	if r.query == nil {
		r.query = url.Values{}
	}
	if r.form == nil {
		r.form = url.Values{}
	}
	return nil
}

// RequestOption define option type for request
type RequestOption func(*request)

// WithRecvWindow set recvWindow param for the request
func WithRecvWindow(recvWindow string) RequestOption {
	return func(r *request) {
		r.nonce = recvWindow
	}
}

// WithHeader set or add a header value to the request
func WithHeader(key, value string, replace bool) RequestOption {
	return func(r *request) {
		if r.header == nil {
			r.header = http.Header{}
		}
		if replace {
			r.header.Set(key, value)
		} else {
			r.header.Add(key, value)
		}
	}
}

// WithHeaders set or replace the headers of the request
func WithHeaders(header http.Header) RequestOption {
	return func(r *request) {
		r.header = header.Clone()
	}
}
