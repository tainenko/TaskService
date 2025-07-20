package response

import "github.com/gin-gonic/gin"

type Result struct {
	Code    ErrCode     `json:"code"`
	Message string      `json:"msg"`
	Status  int         `json:"-"`
	Data    interface{} `json:"data"`
}

func (r *Result) Error() string {
	return r.Message
}

func New(code ErrCode, message string, data interface{}) *Result {
	return &Result{
		Code:    code,
		Message: message,
		Status:  200,
		Data:    data,
	}
}
func (r *Result) WithError(status int, code ErrCode, message string) *Result {
	r.Status = status
	r.Code = code
	r.Message = message
	return r
}

func (r *Result) WithData(data interface{}) *Result {
	r.Data = data
	return r
}

func Fail(c *gin.Context, status int, code ErrCode, message string) {
	c.JSON(status, Result{
		Code:    code,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Result{
		Code:    0,
		Message: "",
		Data:    data,
	})
}
