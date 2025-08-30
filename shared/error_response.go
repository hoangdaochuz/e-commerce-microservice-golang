package shared

type ErrorType string

const (
	Not_Found           ErrorType = "Not_Found"
	Internal_Server_Err ErrorType = "Internal_Server_Err"
)

type ErrorResponse struct {
	Err        string
	StatusCode int
}
