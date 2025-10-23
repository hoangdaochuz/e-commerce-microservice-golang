package shared

type ContextKey string

const (
	HTTPRequest_ContextKey  ContextKey = "httpRequest"
	HTTPResponse_ContextKey ContextKey = "httpResponse"
	UserId_ContextKey       ContextKey = "userId"
)
