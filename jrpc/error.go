package jsonrpc

// Error object
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

// ErrParseError Invalid JSON was received by the server.
// An error occurred on the server while parsing the JSON text.
func ErrParseError(data any) *Error {
	return &Error{
		Code:    parseError.Int(),
		Message: parseError.String(),
		Data:    data,
	}
}

// ErrInvalidRequest The JSON sent is not a valid Request object.
func ErrInvalidRequest(data any) *Error {
	return &Error{
		Code:    invalidRequest.Int(),
		Message: invalidRequest.String(),
		Data:    data,
	}
}

// ErrMethodNotFound The method does not exist / is not available.
func ErrMethodNotFound(data any) *Error {
	return &Error{
		Code:    methodNotFound.Int(),
		Message: methodNotFound.String(),
		Data:    data,
	}
}

// ErrInvalidParams Invalid method parameter(s).
func ErrInvalidParams(data any) *Error {
	return &Error{
		Code:    invalidParams.Int(),
		Message: invalidParams.String(),
		Data:    data,
	}
}

// ErrInternalError Internal JSON-RPC error.
func ErrInternalError(data any) *Error {
	return &Error{
		Code:    internalError.Int(),
		Message: internalError.String(),
		Data:    data,
	}
}
