package api

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"

	"bin-vul-inspector/app/kit"
	_ "bin-vul-inspector/docs"
	"bin-vul-inspector/pkg/api/v1"
)

func NewHttpHandler(kit *kit.Kit) http.Handler {
	// csrfMiddleware := CsrfMiddleware(kit)
	// csrfTokenMiddleware := CsrfTokenMiddleware()

	gin.DisableConsoleColor()
	gin.DisableBindValidation()

	router := gin.Default()
	router.UseH2C = true
	router.MaxMultipartMemory = 20 << 20 // 20M
	router.Use(requestid.New(), cors.Default())

	base := v1.NewBase(kit)

	router.GET("/ping", base.Ping)
	router.GET("/version", base.Version)

	scsRouter := router.Group("/scs")
	{
		// swagger
		if kit.Config.DebugMode {
			scsRouter.GET("/swagger/*any", gin.WrapF(httpSwagger.Handler()))
		}
	}

	v1Router := scsRouter.Group("/api/v1")

	// task
	{
		taskHandler := v1.NewTask(base)

		v1Router.Group("/tasks").
			POST("", taskHandler.Create).
			GET("", taskHandler.List)
		v1Router.Group("/tasks/:task_id").
			GET("", taskHandler.Detail).
			DELETE("", taskHandler.Delete).
			POST("/terminate", taskHandler.Terminate).
			GET("/:type/log", taskHandler.LogFile).
			GET("/:type/asm_file", taskHandler.ASMFile)
	}

	// bha
	{
		bhaHandler := v1.NewBha(base)

		// task
		v1Router.Group("/bha/task").
			GET("/:task_id/file/funcs", bhaHandler.ListFunc).
			GET("/:task_id/file/func_results", bhaHandler.ListFuncResult).
			GET("/:task_id/report", bhaHandler.GetReport)

		// model
		v1Router.Group("/bha/model").
			POST("", bhaHandler.UploadModel).
			GET("", bhaHandler.ListModel).
			GET("/:model_id", bhaHandler.DetailModel).
			DELETE("/:model_id", bhaHandler.DeleteModel)
	}

	// 调试模式时
	if kit.Config.DebugMode {
		// pprof
		pprof.Register(router)
		// 转发所有未匹配请求。
		if kit.Config.DevProxy != "" {
			router.NoRoute(ReverseProxy(kit.Config.DevProxy))
		}
	}

	return router.Handler()
}

func ReverseProxy(target string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		remote, err := url.Parse(target)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = ctx.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = ctx.Request.URL.Path
		}
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
