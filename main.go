package main

import (
	"github.com/gin-gonic/gin"
	"github.com/smilelinkd/bowexecutor/device"
	"github.com/smilelinkd/bowexecutor/router"
	"github.com/smilelinkd/bowexecutor/service"
	"k8s.io/klog/v2"
)

func Setup() *gin.Engine {
	service.SocketInit()
	client, err := device.InitBow()
	if err != nil {
		klog.Errorf("Init error: %v", err)
		return nil
	}
	return router.NewRouter(client)
}

func main() {
	r := Setup()
	if err := r.Run(":8080"); err != nil {
		klog.Errorf("Run start error: %v", err)
	}
}
