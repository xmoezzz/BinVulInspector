package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/gwatts/gin-adapter"

	"bin-vul-inspector/app/kit"
	v1 "bin-vul-inspector/pkg/api/v1"
)

func CsrfMiddleware(kit *kit.Kit) gin.HandlerFunc {
	return func(c *gin.Context) {
		csrfProtect := csrf.Protect(
			[]byte("zsZHCNPvSxBhXFdpfcHwvd7q2pVTYe8J"),
			csrf.CookieName("bin-vul-inspector"),
			csrf.FieldName("_token"),
			csrf.RequestHeader("X-Token"),
			csrf.Path("/"),
			csrf.Secure(false),
			csrf.ErrorHandler(http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					v1.NewBase(kit).ForbiddenMsg(c, "token错误")
				},
			)),
		)
		adapter.Wrap(csrfProtect)(c)
	}
}

func CsrfTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := csrf.Token(ctx.Request)
		ctx.Header("X-Token", token)
	}
}
