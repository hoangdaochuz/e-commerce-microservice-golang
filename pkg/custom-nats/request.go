package custom_nats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	backend_endpont_key = "general_config.backend_endpoint"
)

func init() {
	viper.SetDefault(backend_endpont_key, "http://localhost:8080")
}

type Request struct {
	Header      map[string][]string
	Method      string
	Body        []byte
	URL         string
	Subject     string
	ServiceName string
}

var specialEnpointMap = map[string]string{
	"callback": "/api/v1/auth/Callback",
}

func NewRequest(header map[string][]string, method, url, subject string, body []byte) *Request {
	serviceName := strings.Split(subject, "/")[3]
	return &Request{
		Header:      header,
		Method:      method,
		Body:        body,
		URL:         url,
		Subject:     subject,
		ServiceName: serviceName,
	}
}

func (r *Request) AddHeader(key, value string) {
	_, ok := r.Header[key]
	if ok {
		r.Header[key] = append(r.Header[key], value)
	} else {
		r.Header[key] = []string{value}
	}
}

func (r *Request) GetServiceName() string {
	return r.ServiceName
}

func buildNatsSubjectFromPath(path string) string {
	splits := strings.Split(path, "/")
	var subject string
	var i = 1
	for i < 3 {
		subject += splits[i] + "/"
		i++
	}
	subject = "/" + subject + splits[i]
	return subject
}

func copyCookieFromHTTPRequest(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}
	cookiesString := []string{}
	for _, cookie := range cookies {
		cookiesString = append(cookiesString, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	return strings.Join(cookiesString, "; ")
}

func convertHttpGetRequestToNatsPostRequest(r http.Request) (*Request, error) {
	urlObject := r.URL
	host := urlObject.Host
	path := urlObject.Path
	if host == "" {
		host = viper.GetString(backend_endpont_key)
	}
	queryParams := urlObject.RawQuery
	querySplits := strings.Split(queryParams, "&")
	queryObj := make(map[string]string)
	for _, query := range querySplits {
		keyValuePair := strings.Split(query, "=")
		queryObj[cases.Title(language.English).String(keyValuePair[0])] = keyValuePair[1]
	}
	body, err := json.Marshal(queryObj)
	if err != nil {
		return nil, err
	}
	splitsPath := strings.Split(path, "/")
	if len(splitsPath) == 2 {
		v, ok := specialEnpointMap[splitsPath[1]]
		if !ok {
			return nil, fmt.Errorf("path is not valid")
		}
		path = v
	}
	subject := buildNatsSubjectFromPath(path)

	serviceName := strings.Split(subject, "/")[3]

	urlString := host + path
	headers := make(map[string][]string)
	for key, values := range r.Header {
		headers[key] = append(headers[key], values...)
	}
	cookie := copyCookieFromHTTPRequest(r.Cookies())
	if cookie != "" {
		headers["Cookie"] = []string{copyCookieFromHTTPRequest(r.Cookies())}
	}

	return &Request{
		Method:      "POST",
		Body:        body,
		Header:      headers,
		Subject:     subject,
		URL:         urlString,
		ServiceName: serviceName,
	}, nil
}

func HttpRequestToNatsRequest(r http.Request) (*Request, error) {
	method := r.Method
	headers := make(map[string][]string)
	if method == "GET" {
		return convertHttpGetRequestToNatsPostRequest(r)
	}
	urlObject := r.URL
	host := urlObject.Host
	path := urlObject.Path
	subject := buildNatsSubjectFromPath(path)
	serviceName := strings.Split(subject, "/")[3]
	if path[0] != '/' {
		path = "/" + path
	}
	urlString := host + path

	bodyReader := r.Body
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	err = bodyReader.Close()
	if err != nil {
		return nil, err
	}

	for headerKey, values := range r.Header {
		headers[headerKey] = append(headers[headerKey], values...)
	}

	cookie := copyCookieFromHTTPRequest(r.Cookies())
	if cookie != "" {
		headers["Cookie"] = []string{cookie}
	}

	return &Request{
		Method:      method,
		URL:         urlString,
		Header:      headers,
		Body:        body,
		Subject:     subject,
		ServiceName: serviceName,
	}, nil
}

func NatsRequestToHttpRequest(rq *Request) (*http.Request, error) {
	method := rq.Method
	url := rq.URL
	bodyByte := rq.Body
	bodyReader := bytes.NewReader(bodyByte)

	httpRequest, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range rq.Header {
		for _, item := range v {
			httpRequest.Header.Add(k, item)
		}
	}
	return httpRequest, nil
}

type RequestBuilder struct {
	Request
}

func (b *RequestBuilder) AddHeader(h map[string][]string) *RequestBuilder {
	b.Header = h
	return b
}

func (b *RequestBuilder) AddMethod(method string) *RequestBuilder {
	b.Method = method
	return b
}
func (b *RequestBuilder) AddSubject(subject string) *RequestBuilder {
	b.Subject = subject
	return b
}

func (b *RequestBuilder) AddBody(body []byte) *RequestBuilder {
	b.Body = body
	return b
}

func (b *RequestBuilder) AddUrl(url string) *RequestBuilder {
	b.URL = url
	return b
}

func (b *RequestBuilder) Build() *Request {
	return &Request{
		Header:  b.Header,
		URL:     b.URL,
		Method:  b.Method,
		Body:    b.Body,
		Subject: b.Subject,
	}
}
