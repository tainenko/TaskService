package response

type ErrCode int

//go:generate stringer -type ErrCode -linecomment

const (
	// Server related errors start from 10001
	ServerError    ErrCode = iota + 10001 // Server Error
	InvalidParam                          // Invalid Param
	InvalidPayload                        // Invalid Payload
)

const (
	// Task related errors start from 20001
	CreateTaskErr ErrCode = iota + 20001 // Insert Task Error
	GetTasksErr                          // Get Tasks Error
	UpdateTaskErr                        // Update Task Error
	DeleteTaskErr                        // Delete Task Error
	GetTaskErr                           // Get Task Error
)

type APIError struct {
	Code   ErrCode `json:"code"`
	Msg    string  `json:"msg"`
	Status int     `json:"status"`
}

func (e *APIError) Error() string {
	return e.Msg
}

func Wrap(status int, err error, code ErrCode) APIError {
	return APIError{
		Code:   code,
		Msg:    err.Error(),
		Status: status,
	}
}
