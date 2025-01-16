package v1

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/cmd/version"
	"bin-vul-inspector/pkg/api/services"
	"bin-vul-inspector/pkg/api/v1/dto"
)

type Base struct {
	*kit.Kit
}

func NewBase(kit *kit.Kit) *Base {
	return &Base{
		Kit: kit,
	}
}

func (h *Base) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    version.Version,
		"commit":     version.GitCommit,
		"build_time": version.BuildTime,
	})
}

func (h *Base) Ping(c *gin.Context) {
	_, _ = c.Writer.Write([]byte("pong\n"))
}

func (h *Base) response(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func (h *Base) File(c *gin.Context, path string) {
	c.File(path)
}

func (h *Base) Success(c *gin.Context, data interface{}) {
	h.response(c, http.StatusOK, dto.Response{Code: dto.StatusOk, Data: data})
}

func (h *Base) resStatusText(c *gin.Context, httpCode int, statusCode int) {
	h.response(c, httpCode, dto.Response{Code: statusCode, ErrMessage: dto.StatusText(statusCode)})
}

func (h *Base) resMsg(c *gin.Context, httpCode int, statusCode int, errMessage string) {
	h.response(c, httpCode, dto.Response{Code: statusCode, ErrMessage: errMessage})
}

func (h *Base) Fail(c *gin.Context, statusCode int) {
	h.resStatusText(c, http.StatusOK, statusCode)
}

func (h *Base) FailMsg(c *gin.Context, statusCode int, errMessage string) {
	h.resMsg(c, http.StatusOK, statusCode, errMessage)
}

func (h *Base) BadRequest(c *gin.Context, statusCode int) {
	h.resStatusText(c, http.StatusBadRequest, statusCode)
}

func (h *Base) BadRequestMsg(c *gin.Context, statusCode int, errMessage string) {
	h.resMsg(c, http.StatusBadRequest, statusCode, errMessage)
}

func (h *Base) ForbiddenMsg(c *gin.Context, errMessage string) {
	h.resMsg(c, http.StatusForbidden, http.StatusForbidden, errMessage)
}

func (h *Base) InternalError(c *gin.Context, statusCode int) {
	h.resStatusText(c, http.StatusInternalServerError, statusCode)
}

func (h *Base) InternalErrorMsg(c *gin.Context, statusCode int, errMessage string) {
	h.resMsg(c, http.StatusInternalServerError, statusCode, errMessage)
}

func (h *Base) Error(ctx *gin.Context, err error) {
	var serviceErr *services.Error
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &serviceErr):
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			dto.Response{
				Code:       serviceErr.Code,
				ErrMessage: serviceErr.Message,
			})
	case errors.As(err, &maxBytesErr):
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity,
			dto.Response{
				Code:       dto.StatusRequestBodyLarge,
				ErrMessage: dto.StatusText(dto.StatusRequestBodyLarge),
			})
	default:
		h.Logger.Errorf("Internal Server Error[%s]: %s, Debug Stack:\n %s", requestid.Get(ctx), err, debug.Stack())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			dto.Response{
				Code:       dto.StatusInternalError,
				ErrMessage: dto.StatusText(dto.StatusInternalError),
			})
	}
}

func (h *Base) ErrorParseFormData(ctx *gin.Context, err error) {
	h.Error(ctx, services.NewError(dto.StatusErrParseFormData, dto.StatusText(dto.StatusErrParseFormData), err))
}

func (h *Base) ErrorParseJson(ctx *gin.Context, err error) {
	h.Error(ctx, services.NewError(dto.StatusErrJson, dto.StatusText(dto.StatusErrJson), err))
}

func (h *Base) ErrorValidate(ctx *gin.Context, err error) {
	h.Error(ctx, services.NewError(dto.StatusParamInvalid, err.Error()))
}
