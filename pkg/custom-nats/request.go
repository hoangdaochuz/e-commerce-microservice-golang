package custom_nats

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	Header  map[string][]string
	Method  string
	Body    []byte
	URL     string
	Subject string
}

func NewRequest(header map[string][]string, method, url, subject string, body []byte) *Request {
	return &Request{
		Header:  header,
		Method:  method,
		Body:    body,
		URL:     url,
		Subject: subject,
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

func HttpRequestToNatsRequest(r http.Request) (*Request, error) {
	method := r.Method
	urlObject := r.URL
	host := urlObject.Host

	path := urlObject.Path
	subject := buildNatsSubjectFromPath(path)
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

	return &Request{
		Method:  method,
		URL:     urlString,
		Header:  r.Header,
		Body:    body,
		Subject: subject,
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
