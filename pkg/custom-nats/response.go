package custom_nats

import "net/http"

type ResponseInterface interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Response struct {
	StatusCode int
	Status     string
	Body       []byte
	Headers    http.Header
}

func NewResponse(statusCode int, headers http.Header, body []byte, status string) *Response {
	return &Response{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
		Status:     status,
	}
}

func (res *Response) GetStatusCode() int {
	return res.StatusCode
}

func (res *Response) SetStatusCode(code int) {
	res.StatusCode = code
}

func (res *Response) GetHeaders() http.Header {
	return res.Headers
}

func (res *Response) SetHeaders(headers http.Header) {
	res.Headers = headers
}

func (res *Response) GetBody() []byte {
	return res.Body
}
func (res *Response) SetBody(body []byte) {
	res.Body = body
}

func (res *Response) GetStatus() string {
	return res.Status
}

func (res *Response) SetStatus(status string) {
	res.Status = status
}

// Implement Header ResponseInterface
func (res *Response) Header() http.Header {
	return res.Headers
}

// Implement Write([]byte) (int, error) ResponseInterface
func (res *Response) Write(body []byte) (int, error) {
	res.Body = body
	return 1, nil
}

// Implement WriteHeader(statusCode int) ResponseInterface
func (res *Response) WriteHeader(statusCode int) {
	res.StatusCode = statusCode
}

type ResponseBuilder struct {
	res *Response
}

func NewResponseBuilder(statusCode int) *ResponseBuilder {
	return &ResponseBuilder{
		res: &Response{
			StatusCode: statusCode,
		},
	}
}

func (b *ResponseBuilder) BuildBody(body []byte) *ResponseBuilder {
	b.res.Body = body
	return b
}

func (b *ResponseBuilder) BuildHeader(header http.Header) *ResponseBuilder {
	b.res.Headers = header
	return b
}

func (b *ResponseBuilder) BuildStatusCode(statusCode int) *ResponseBuilder {
	b.res.StatusCode = statusCode
	return b
}

func (b *ResponseBuilder) Build() *Response {
	return b.res
}
