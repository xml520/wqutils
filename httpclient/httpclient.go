// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

// Powerful and easy to use http client
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"strings"

	"time"

	"io"
	"io/ioutil"
	"sync"

	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"

	"crypto/tls"

	"compress/gzip"

	"encoding/json"
	"mime/multipart"
)

// Constants definations
// CURL options, see https://github.com/bagder/curl/blob/169fedbdce93ecf14befb6e0e1ce6a2d480252a3/packages/OS400/curl.inc.in
const (
	VERSION   = "0.6.9"
	USERAGENT = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
)

const (
	PROXY_HTTP int = iota
	PROXY_SOCKS4
	PROXY_SOCKS5
	PROXY_SOCKS4A

	// CURL like OPT
	OPT_AUTOREFERER
	OPT_FOLLOWLOCATION
	OPT_CONNECTTIMEOUT
	OPT_CONNECTTIMEOUT_MS
	OPT_MAXREDIRS
	OPT_PROXYTYPE
	OPT_TIMEOUT
	OPT_TIMEOUT_MS
	OPT_COOKIEJAR
	OPT_INTERFACE
	OPT_PROXY
	OPT_REFERER
	OPT_USERAGENT

	// Other OPT
	OPT_REDIRECT_POLICY
	OPT_PROXY_FUNC
	OPT_DEBUG
	OPT_UNSAFE_TLS

	OPT_CONTEXT

	OPT_BEFORE_REQUEST_FUNC
	OPT_AFTER_REQUEST_FUNC

	OPT_SELECT_IP
)

// String map of options
var CONST = map[string]int{
	"OPT_AUTOREFERER":         OPT_AUTOREFERER,
	"OPT_FOLLOWLOCATION":      OPT_FOLLOWLOCATION,
	"OPT_CONNECTTIMEOUT":      OPT_CONNECTTIMEOUT,
	"OPT_CONNECTTIMEOUT_MS":   OPT_CONNECTTIMEOUT_MS,
	"OPT_MAXREDIRS":           OPT_MAXREDIRS,
	"OPT_PROXYTYPE":           OPT_PROXYTYPE,
	"OPT_TIMEOUT":             OPT_TIMEOUT,
	"OPT_TIMEOUT_MS":          OPT_TIMEOUT_MS,
	"OPT_COOKIEJAR":           OPT_COOKIEJAR,
	"OPT_INTERFACE":           OPT_INTERFACE,
	"OPT_PROXY":               OPT_PROXY,
	"OPT_REFERER":             OPT_REFERER,
	"OPT_USERAGENT":           OPT_USERAGENT,
	"OPT_REDIRECT_POLICY":     OPT_REDIRECT_POLICY,
	"OPT_PROXY_FUNC":          OPT_PROXY_FUNC,
	"OPT_DEBUG":               OPT_DEBUG,
	"OPT_UNSAFE_TLS":          OPT_UNSAFE_TLS,
	"OPT_CONTEXT":             OPT_CONTEXT,
	"OPT_BEFORE_REQUEST_FUNC": OPT_BEFORE_REQUEST_FUNC,
	"OPT_AFTER_REQUEST_FUNC":  OPT_AFTER_REQUEST_FUNC,
	"OPT_SELECT_IP":           OPT_SELECT_IP,
}

// Default options for any clients.
var defaultOptions = map[int]interface{}{
	OPT_FOLLOWLOCATION: true,
	OPT_MAXREDIRS:      10,
	OPT_AUTOREFERER:    true,
	OPT_USERAGENT:      USERAGENT,
	OPT_COOKIEJAR:      true,
	OPT_DEBUG:          false,
}

// These options affect transport, transport may not be reused if you change any
// of these options during a request.
var transportOptions = []int{
	OPT_CONNECTTIMEOUT,
	OPT_CONNECTTIMEOUT_MS,
	OPT_PROXYTYPE,
	OPT_TIMEOUT,
	OPT_TIMEOUT_MS,
	OPT_INTERFACE,
	OPT_PROXY,
	OPT_PROXY_FUNC,
	OPT_UNSAFE_TLS,
}

// These options affect cookie jar, jar may not be reused if you change any of
// these options during a request.
var jarOptions = []int{
	OPT_COOKIEJAR,
}

// Thin wrapper of http.Response(can also be used as http.Response).
type Response struct {
	*http.Response
	body []byte
}

// Read response body into a byte slice.
func (res *Response) ReadAll() ([]byte, error) {
	if res.body != nil {
		return res.body, nil
	}
	var reader io.ReadCloser
	var err error

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
	default:
		reader = res.Body
	}
	defer reader.Close()
	res.body, err = ioutil.ReadAll(reader)
	return res.body, err
}

// Read response body into string.
func (res *Response) ToString() string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ToString", err)
		}
	}()
	bytes, err := res.ReadAll()
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (res *Response) ToJson(path string) gjson.Result {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ToString", err)
		}
	}()
	bytes, err := res.ReadAll()
	if err != nil {
		return gjson.Result{}
	}
	if path == "" {
		return gjson.ParseBytes(bytes)
	} else {
		return gjson.GetBytes(bytes, path)
	}
}

// Prepare a request.
func prepareRequest(method string, url_ string, headers map[string]string,
	body io.Reader, options map[int]interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, url_, body)

	if err != nil {
		return nil, err
	}

	// OPT_REFERER
	if referer, ok := options[OPT_REFERER]; ok {
		if refererStr, ok := referer.(string); ok {
			req.Header.Set("Referer", refererStr)
		}
	}

	// OPT_USERAGENT
	if useragent, ok := options[OPT_USERAGENT]; ok {
		if useragentStr, ok := useragent.(string); ok {
			req.Header.Set("User-Agent", useragentStr)
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// Prepare a transport.
//
// Handles timemout, proxy and maybe other transport related options here.
func prepareTransport(options map[int]interface{}) (http.RoundTripper, error) {
	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	var connectTimeout time.Duration

	if connectTimeoutMS_, ok := options[OPT_CONNECTTIMEOUT_MS]; ok {
		if connectTimeoutMS, ok := connectTimeoutMS_.(int); ok {
			connectTimeout = time.Duration(connectTimeoutMS) * time.Millisecond
		} else {
			return nil, fmt.Errorf("OPT_CONNECTTIMEOUT_MS must be int")
		}
	} else if connectTimeout_, ok := options[OPT_CONNECTTIMEOUT]; ok {
		if connectTimeout, ok = connectTimeout_.(time.Duration); !ok {
			if connectTimeoutS, ok := connectTimeout_.(int); ok {
				connectTimeout = time.Duration(connectTimeoutS) * time.Second
			} else {
				return nil, fmt.Errorf("OPT_CONNECTTIMEOUT must be int or time.Duration")
			}
		}
	}

	var timeout time.Duration

	if timeoutMS_, ok := options[OPT_TIMEOUT_MS]; ok {
		if timeoutMS, ok := timeoutMS_.(int); ok {
			timeout = time.Duration(timeoutMS) * time.Millisecond
		} else {
			return nil, fmt.Errorf("OPT_TIMEOUT_MS must be int")
		}
	} else if timeout_, ok := options[OPT_TIMEOUT]; ok {
		if timeout, ok = timeout_.(time.Duration); !ok {
			if timeoutS, ok := timeout_.(int); ok {
				timeout = time.Duration(timeoutS) * time.Second
			} else {
				return nil, fmt.Errorf("OPT_TIMEOUT must be int or time.Duration")
			}
		}
	}

	// fix connect timeout(important, or it might cause a long time wait during
	//connection)
	if timeout > 0 && (connectTimeout > timeout || connectTimeout == 0) {
		connectTimeout = timeout
	}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var conn net.Conn
		var err error
		var lAddr *net.TCPAddr
		var rAddr *net.TCPAddr
		if ip, ok := options[OPT_SELECT_IP]; ok && ip.(string) != "" {
			lAddr, err = net.ResolveTCPAddr(network, ip.(string)+":0")
			if err != nil {
				return nil, err
			}
			rAddr, err = net.ResolveTCPAddr(network, addr)
			if err != nil {
				return nil, err
			}

			conn, err = net.DialTCP(network, lAddr, rAddr)
			if err != nil {
				return nil, err
			}
			if timeout > 0 {
				conn.SetDeadline(time.Now().Add(timeout))
			}
			return conn, err
		}
		if connectTimeout > 0 {
			conn, err = net.DialTimeout(network, addr, connectTimeout)

			if err != nil {
				return nil, err
			}
		} else {
			//fmt.Println(addr, timeout)
			conn, err = net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
		}

		if timeout > 0 {
			conn.SetDeadline(time.Now().Add(timeout))
		}

		return conn, nil
	}

	// proxy
	if proxyFunc_, ok := options[OPT_PROXY_FUNC]; ok {
		if proxyFunc, ok := proxyFunc_.(func(*http.Request) (int, string, error)); ok {
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				proxyType, u_, err := proxyFunc(req)
				if err != nil {
					return nil, err
				}

				if proxyType != PROXY_HTTP {
					return nil, fmt.Errorf("only PROXY_HTTP is currently supported")
				}

				u_ = "http://" + u_

				u, err := url.Parse(u_)

				if err != nil {
					return nil, err
				}

				return u, nil
			}
		} else {
			return nil, fmt.Errorf("OPT_PROXY_FUNC is not a desired function")
		}
	} else {
		var proxytype int
		if proxytype_, ok := options[OPT_PROXYTYPE]; ok {
			if proxytype, ok = proxytype_.(int); !ok || proxytype != PROXY_HTTP {
				return nil, fmt.Errorf("OPT_PROXYTYPE must be int, and only PROXY_HTTP is currently supported")
			}
		}

		var proxy string
		if proxy_, ok := options[OPT_PROXY]; ok {
			if proxy, ok = proxy_.(string); !ok {
				return nil, fmt.Errorf("OPT_PROXY must be string")
			}

			if !strings.Contains(proxy, "://") {
				proxy = "http://" + proxy
			}
			proxyUrl, err := url.Parse(proxy)
			if err != nil {
				return nil, err
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	// TLS
	if unsafe_tls_, found := options[OPT_UNSAFE_TLS]; found {
		var unsafe_tls, _ = unsafe_tls_.(bool)
		var tls_config = transport.TLSClientConfig
		if tls_config == nil {
			tls_config = &tls.Config{}
			transport.TLSClientConfig = tls_config
		}
		tls_config.InsecureSkipVerify = unsafe_tls

	}

	return transport, nil
}

// Prepare a redirect policy.
func prepareRedirect(options map[int]interface{}) (func(req *http.Request, via []*http.Request) error, error) {
	var redirectPolicy func(req *http.Request, via []*http.Request) error

	if redirectPolicy_, ok := options[OPT_REDIRECT_POLICY]; ok {
		if redirectPolicy, ok = redirectPolicy_.(func(*http.Request, []*http.Request) error); !ok {
			return nil, fmt.Errorf("OPT_REDIRECT_POLICY is not a desired function")
		}
	} else {
		var followlocation bool
		if followlocation_, ok := options[OPT_FOLLOWLOCATION]; ok {
			if followlocation, ok = followlocation_.(bool); !ok {
				return nil, fmt.Errorf("OPT_FOLLOWLOCATION must be bool")
			}
		}

		var maxredirs int
		if maxredirs_, ok := options[OPT_MAXREDIRS]; ok {
			if maxredirs, ok = maxredirs_.(int); !ok {
				return nil, fmt.Errorf("OPT_MAXREDIRS must be int")
			}
		}

		redirectPolicy = func(req *http.Request, via []*http.Request) error {
			// no follow
			if !followlocation || maxredirs <= 0 {
				return &Error{
					Code:    ERR_REDIRECT_POLICY,
					Message: fmt.Sprintf("redirect not allowed"),
				}
			}

			if len(via) >= maxredirs {
				return &Error{
					Code:    ERR_REDIRECT_POLICY,
					Message: fmt.Sprintf("stopped after %d redirects", len(via)),
				}
			}

			last := via[len(via)-1]
			// keep necessary headers
			// TODO: pass all headers or add other headers?
			if useragent := last.Header.Get("User-Agent"); useragent != "" {
				req.Header.Set("User-Agent", useragent)
			}

			return nil
		}
	}

	return redirectPolicy, nil
}

// Prepare a cookie jar.
func prepareJar(options map[int]interface{}) (http.CookieJar, error) {
	var jar http.CookieJar
	var err error
	if optCookieJar_, ok := options[OPT_COOKIEJAR]; ok {
		// is bool
		if optCookieJar, ok := optCookieJar_.(bool); ok {
			// default jar
			if optCookieJar {
				// TODO: PublicSuffixList
				jar, err = cookiejar.New(nil)
				if err != nil {
					return nil, err
				}
			}
		} else if optCookieJar, ok := optCookieJar_.(http.CookieJar); ok {
			jar = optCookieJar
		} else {
			return nil, fmt.Errorf("invalid cookiejar")
		}
	}

	return jar, nil
}

// Create an HTTP client.
func NewHttpClient() *HttpClient {
	c := &HttpClient{
		reuseTransport: true,
		reuseJar:       true,
		lock:           new(sync.Mutex),
	}

	return c
}

// Powerful and easy to use HTTP client.
type HttpClient struct {
	// Default options of this client.
	options map[int]interface{}

	// Default headers of this client.
	Headers map[string]string

	// Options of current request.
	oneTimeOptions map[int]interface{}

	// Headers of current request.
	oneTimeHeaders map[string]string

	// Cookies of current request.
	oneTimeCookies []*http.Cookie

	// Global transport of this client, might be shared between different
	// requests.
	transport http.RoundTripper

	// Global cookie jar of this client, might be shared between different
	// requests.
	jar http.CookieJar

	// Whether current request should reuse the transport or not.
	reuseTransport bool

	// Whether current request should reuse the cookie jar or not.
	reuseJar bool

	// Make requests of one client concurrent safe.
	lock *sync.Mutex

	withLock bool
}

// Set default options and headers.
func (h *HttpClient) Defaults(defaults Map) *HttpClient {
	options, headers := parseMap(defaults)

	// merge options
	if h.options == nil {
		h.options = options
	} else {
		for k, v := range options {
			h.options[k] = v
		}
	}

	// merge headers
	if h.Headers == nil {
		h.Headers = headers
	} else {
		for k, v := range headers {
			h.Headers[k] = v
		}
	}

	return h
}

// Begin marks the begining of a request, it's necessary for concurrent
// requests.
func (h *HttpClient) Begin() *HttpClient {
	h.lock.Lock()
	h.withLock = true

	return h
}

// Reset the client state so that other requests can begin.
func (h *HttpClient) reset() {
	h.oneTimeOptions = nil
	h.oneTimeHeaders = nil
	h.oneTimeCookies = nil
	h.reuseTransport = true
	h.reuseJar = true

	// nil means the Begin has not been called, asume requests are not
	// concurrent.
	if h.withLock {
		h.withLock = false
		h.lock.Unlock()
	}
}

// Temporarily specify an option of the current request.
func (h *HttpClient) WithOption(k int, v interface{}) *HttpClient {
	if h.oneTimeOptions == nil {
		h.oneTimeOptions = make(map[int]interface{})
	}
	h.oneTimeOptions[k] = v

	// Conditions we cann't reuse the transport.
	if hasOption(k, transportOptions) {
		h.reuseTransport = false
	}

	// Conditions we cann't reuse the cookie jar.
	if hasOption(k, jarOptions) {
		h.reuseJar = false
	}

	return h
}

// Temporarily specify multiple options of the current request.
func (h *HttpClient) WithOptions(m Map) *HttpClient {
	options, _ := parseMap(m)
	for k, v := range options {
		h.WithOption(k, v)
	}

	return h
}

// Temporarily specify a header of the current request.
func (h *HttpClient) WithHeader(k string, v string) *HttpClient {
	if h.oneTimeHeaders == nil {
		h.oneTimeHeaders = make(map[string]string)
	}
	h.oneTimeHeaders[k] = v

	return h
}

// Temporarily specify multiple headers of the current request.
func (h *HttpClient) WithHeaders(m map[string]string) *HttpClient {
	for k, v := range m {
		h.WithHeader(k, v)
	}

	return h
}

// Specify cookies of the current request.
func (h *HttpClient) WithCookie(cookies ...*http.Cookie) *HttpClient {
	h.oneTimeCookies = append(h.oneTimeCookies, cookies...)

	return h
}

// Start a request, and get the response.
//
// Usually we just need the Get and Post method.
func (h *HttpClient) Do(method string, url string, headers map[string]string,
	body io.Reader) (*Response, error) {
	options := mergeOptions(defaultOptions, h.options, h.oneTimeOptions)
	headers = mergeHeaders(h.Headers, headers, h.oneTimeHeaders)
	cookies := h.oneTimeCookies

	var transport http.RoundTripper
	var jar http.CookieJar
	var err error

	// transport
	if h.transport == nil || !h.reuseTransport {
		transport, err = prepareTransport(options)
		if err != nil {
			h.reset()
			return nil, err
		}

		if h.reuseTransport {
			h.transport = transport
		}
	} else {
		transport = h.transport
	}

	// jar
	if h.jar == nil || !h.reuseJar {
		jar, err = prepareJar(options)
		if err != nil {
			h.reset()
			return nil, err
		}

		if h.reuseJar {
			h.jar = jar
		}
	} else {
		jar = h.jar
	}

	// release lock
	h.reset()

	redirect, err := prepareRedirect(options)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Transport:     transport,
		CheckRedirect: redirect,
		Jar:           jar,
	}

	req, err := prepareRequest(method, url, headers, body, options)
	if err != nil {
		return nil, err
	}
	if debugEnabled, ok := options[OPT_DEBUG]; ok {
		if debugEnabled.(bool) {
			dump, err := httputil.DumpRequestOut(req, true)
			if err == nil {
				fmt.Printf("%s\n", dump)
			}
		}
	}

	if jar != nil {
		jar.SetCookies(req.URL, cookies)
	} else {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	if ctx, ok := options[OPT_CONTEXT]; ok {
		if c, ok := ctx.(context.Context); ok {
			req = req.WithContext(c)
		}
	}

	if beforeReqFunc, ok := options[OPT_BEFORE_REQUEST_FUNC]; ok {
		if f, ok := beforeReqFunc.(func(c *http.Client, r *http.Request)); ok {
			f(c, req)
		}
	}

	res, err := c.Do(req)
	hRes := &Response{res, nil}
	if err != nil {
		return hRes, err
	}
	if beforeReqFunc, ok := options[OPT_AFTER_REQUEST_FUNC]; ok {
		if f, ok := beforeReqFunc.(func(r *Response) error); ok {
			if err = f(hRes); err != nil {
				return hRes, err
			}
		}
	}

	return hRes, err
}

// The HEAD request
func (h *HttpClient) Head(url string) (*Response, error) {
	return h.Do("HEAD", url, nil, nil)
}

// The GET request
func (h *HttpClient) Get(url string, params ...interface{}) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("GET", url, nil, nil)
}

// The DELETE request
func (h *HttpClient) Delete(url string, params ...interface{}) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("DELETE", url, nil, nil)
}

// The POST request
//
// With multipart set to true, the request will be encoded as
// "multipart/form-data".
//
// If any of the params key starts with "@", it is considered as a form file
// (similar to CURL but different).
func (h *HttpClient) Post(url string, params interface{}) (*Response, error) {
	t := checkParamsType(params)
	if t == 2 {
		return h.Do("POST", url, nil, toReader(params))
	}

	paramsValues := toUrlValues(params)
	// Post with files should be sent as multipart.
	if checkParamFile(paramsValues) {
		return h.PostMultipart(url, params)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	body := strings.NewReader(paramsValues.Encode())

	return h.Do("POST", url, headers, body)
}

// Post with the request encoded as "multipart/form-data".
func (h *HttpClient) PostMultipart(url string, params interface{}) (*Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	paramsValues := toUrlValues(params)
	// check files
	for k, v := range paramsValues {
		for _, vv := range v {
			// is file
			if k[0] == '@' {
				err := addFormFile(writer, k[1:], vv)
				if err != nil {
					return nil, err
				}
			} else {
				writer.WriteField(k, vv)
			}
		}
	}
	headers := make(map[string]string)

	headers["Content-Type"] = writer.FormDataContentType()
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return h.Do("POST", url, headers, body)
}

func (h *HttpClient) sendJson(method string, url string, data interface{}) (*Response, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	var body []byte
	switch t := data.(type) {
	case []byte:
		body = t
	case string:
		body = []byte(t)
	default:
		var err error
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	return h.Do(method, url, headers, bytes.NewReader(body))
}

func (h *HttpClient) PostJson(url string, data interface{}) (*Response, error) {
	return h.sendJson("POST", url, data)
}

// The PUT request
func (h *HttpClient) Put(url string, body io.Reader) (*Response, error) {
	return h.Do("PUT", url, nil, body)
}

// Put json data
func (h *HttpClient) PutJson(url string, data interface{}) (*Response, error) {
	return h.sendJson("PUT", url, data)
}

// do json data
func (h *HttpClient) Json(method, url string, data interface{}) (*Response, error) {
	return h.sendJson(method, url, data)
}

// Put json data
func (h *HttpClient) DeleteJson(url string, data interface{}) (*Response, error) {
	return h.sendJson("DELETE", url, data)
}

// Patch json data
func (h *HttpClient) PatchJson(url string, data interface{}) (*Response, error) {
	return h.sendJson("PATCH", url, data)
}

// The OPTIONS request
func (h *HttpClient) Options(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("OPTIONS", url, nil, nil)
}

// The CONNECT request
func (h *HttpClient) Connect(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("CONNECT", url, nil, nil)
}

// The TRACE request
func (h *HttpClient) Trace(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("TRACE", url, nil, nil)
}

// The PATCH request
func (h *HttpClient) Patch(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return h.Do("PATCH", url, nil, nil)
}

// Get cookies of the client jar.
func (h *HttpClient) Cookies(url_ string) []*http.Cookie {
	if h.jar != nil {
		u, _ := url.Parse(url_)
		return h.jar.Cookies(u)
	}

	return nil
}

// Get cookie values(k-v map) of the client jar.
func (h *HttpClient) CookieValues(url_ string) map[string]string {
	m := make(map[string]string)

	for _, c := range h.Cookies(url_) {
		m[c.Name] = c.Value
	}

	return m
}

// Get cookie value of a specified cookie name.
func (h *HttpClient) CookieValue(url_ string, key string) string {
	for _, c := range h.Cookies(url_) {
		if c.Name == key {
			return c.Value
		}
	}

	return ""
}
