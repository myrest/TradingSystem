package bingx

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
	method     string
	endpoint   string
	query      url.Values
	form       url.Values
	recvWindow int64
	header     http.Header
	body       io.Reader
	fullUrl    string
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

// setParam set param with key/value to query string
func (r *request) setParam(key string, value interface{}) *request {
	if r.query == nil {
		r.query = url.Values{}
	}

	if reflect.TypeOf(value).Kind() == reflect.Slice {
		v, err := json.Marshal(value)
		if err == nil {
			value = string(v)
		}
	}

	r.query.Set(key, fmt.Sprintf("%v", value))
	return r
}

func (r *request) setFormParam(key string, value interface{}) *request {
	if r.form == nil {
		r.form = url.Values{}
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
func WithRecvWindow(recvWindow int64) RequestOption {
	return func(r *request) {
		r.recvWindow = recvWindow
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
