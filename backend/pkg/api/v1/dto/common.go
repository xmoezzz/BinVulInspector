package dto

import (
	"errors"
)

const (
	DefaultPageSize = 20
)

type PageParam struct {
	Page     int64 `json:"page" form:"page"`
	PageSize int64 `json:"page_size" form:"page_size"`
}

func (req *PageParam) Validate() error {
	if req.Page < 1 {
		return errors.New("最小页码值为1")
	}
	if req.PageSize < 1 {
		return errors.New("最小页大小为1")
	}
	return nil
}

func (req *PageParam) Skip() int64 {
	return (req.Page - 1) * req.PageSize
}

type Response struct {
	Code       int         `json:"code"`
	Data       interface{} `json:"data"`
	ErrMessage string      `json:"err_message"`
}

type ListResponse[T any] struct {
	Count int64 `json:"count"`
	List  []T   `json:"list"`
}

type CreateRes struct {
	Id string `json:"id"`
}
