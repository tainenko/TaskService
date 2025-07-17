package error

type ErrCode int

//go:generate stringer -type ErrCode -linecomment

const (
	// Server related errors start from 10001
	ServerError ErrCode = iota + 10001
	InvalidParam
	InvalidPayload
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
	Code ErrCode `json:"code"`
	Msg  string  `json:"msg"`
}

func (e *APIError) Error() string {
	return e.Msg
}

func Wrap(err error, code ErrCode) APIError {
	return APIError{
		Code: code,
		Msg:  err.Error(),
	}
}
