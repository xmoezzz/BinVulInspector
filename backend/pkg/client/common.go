package client

const (
	StatusOk = 0
)

type Response struct {
	Code       int    `json:"code"`
	ErrMessage string `json:"err_message"`
}
