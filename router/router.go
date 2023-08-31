package router

import (
	"github.com/smilelinkd/bowexecutor/driver"
	"github.com/smilelinkd/bowexecutor/middleware"
	"github.com/smilelinkd/bowexecutor/pkg/httpadapter"

	"github.com/gin-gonic/gin"
)

func NewRouter(client *driver.DigitalbowClient) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.Cors())
	restController := httpadapter.NewRestController(client)

	route := r.Group("/")
	{
		route.GET("ping", httpadapter.Ping)
		route.POST("download", restController.Download)
		route.GET("websocket", restController.Socket)
	}

	return r
}
