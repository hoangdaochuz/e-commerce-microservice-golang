package shared

var StatusCodeMap = map[ErrorType]int{
	Not_Found:           404,
	Internal_Server_Err: 500,
}
