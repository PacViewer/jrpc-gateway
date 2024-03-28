package jsonrpc

type errorCode int

const (
	parseError     errorCode = -32700
	invalidRequest errorCode = -32600
	methodNotFound errorCode = -32601
	invalidParams  errorCode = -32602
	internalError  errorCode = -32603
	serverError    errorCode = -32000
)

func (e errorCode) Int() int {
	return int(e)
}

func (e errorCode) String() string {
	messages := map[int]string{
		parseError.Int():     "Parse error",
		invalidRequest.Int(): "Invalid Request",
		methodNotFound.Int(): "Method not found",
		invalidParams.Int():  "Invalid params",
		internalError.Int():  "Internal error",
		serverError.Int():    "Server error",
	}

	message, exist := messages[e.Int()]
	if !exist {
		return ""
	}

	return message
}
