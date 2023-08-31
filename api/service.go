package api

//import (
//	"github.com/smilelinkd/bowexecutor/pkg/common"
//	"github.com/smilelinkd/bowexecutor/service"
//	"github.com/smilelinkd/bowexecutor/utils"
//
//	"github.com/gin-gonic/gin"
//)
//
//func Socket(c *gin.Context) {
//	var executeRequest common.ExecuteRequest
//	if err := c.ShouldBind(&executeRequest); err != nil {
//		utils.Warning(c, utils.CodeParamError, "Bad request, failed to decode JSON")
//		return
//	}
//	service.SocketServer(c.Writer, c.Request, executeRequest.ID)
//}
