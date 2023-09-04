package httpadapter

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"gonum.org/v1/gonum/mat"

	"github.com/gin-gonic/gin"
	"github.com/jacobsa/go-serial/serial"
	"github.com/smilelinkd/bowexecutor/pkg/common"
	"github.com/smilelinkd/bowexecutor/service"
	"github.com/smilelinkd/bowexecutor/utils"
)

func (c *RestController) Download(ctx *gin.Context) {
	if c.Client.GetStatus() != common.StatusReady {
		utils.Warning(ctx, utils.CodeNotReady, "For now device is not ready please try next time!")
		return
	}
	response := "This is API " + common.APIVersion + ". Now is " + time.Now().Format(time.UnixDate)
	var downResultRequest common.DownloadRequest
	if err := ctx.ShouldBind(&downResultRequest); err != nil {
		utils.Warning(ctx, utils.CodeParamError, "Bad request, failed to decode JSON")
		return
	}

	c.Client.SetStatus(common.StatusSyncing)
	err := c.Client.DownloadResult(downResultRequest.Path, downResultRequest.Segment)
	if err != nil {
		log.Printf("Can't download file into memory: ", err)
		utils.Warning(ctx, utils.CodeParamError, "Can't download file into memory")
		c.Client.SetStatus(common.StatusReady)
		return
	}
	c.Client.SetStatus(common.StatusReady)
	utils.Success(ctx, nil, response)
}

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (c *RestController) Socket(ctx *gin.Context) {
	var executeRequest common.ExecuteRequest
	if err := ctx.BindQuery(&executeRequest); err != nil {
		utils.Warning(ctx, utils.CodeParamError, "Bad request, failed to decode JSON")
		return
	}
	c.Execute(ctx, executeRequest)
}

func (c *RestController) Execute(ctx *gin.Context, executeRequest common.ExecuteRequest) {
	if c.Client.GetStatus() != common.StatusReady {
		utils.Warning(ctx, utils.CodeNotReady, "For now device is not ready please try next time!")
		return
	}

	if _, ok := c.Client.Movements[executeRequest.Segment]; !ok {
		utils.Warning(ctx, utils.CodeNotReady, "The segment does not exist, please download first!")
		return
	}

	service.SocketServer(ctx.Writer, ctx.Request, executeRequest.ID)

	c.Client.SetStatus(common.StatusExecucting)

	options := serial.OpenOptions{
		PortName:        c.Client.Client.Config.SerialName,
		BaudRate:        uint(c.Client.Client.Config.BaudRate),
		DataBits:        uint(c.Client.Client.Config.DataBits),
		StopBits:        uint(c.Client.Client.Config.StopBits),
		ParityMode:      serial.PARITY_NONE,
		MinimumReadSize: 4,
	}

	go func() {
		port, err := serial.Open(options)
		if err != nil {
			log.Printf("Error opening serial port... %v", err)
			return
		}
		defer port.Close()
		defer c.Client.SetStatus(common.StatusReady)
		trackData := c.Client.Movements[executeRequest.Segment]
		aInit := make([]float64, 6)
		clylen := make([]float32, 6)
		for record, item := range trackData.MatrixList {
			bowResult := c.Client.GetBowDataformat(item, trackData.MatrixInit)
			if record == 0 {
				//# 记录第0帧的初始参数
				aInit = bowResult
			}
			if record < executeRequest.Start { // set start point.
				continue
			} else if record > executeRequest.End { // set end point.
				break
			}
			// send back current index record.
			service.SendClientSocket(executeRequest.ID, strconv.Itoa(record))
			matrixA := mat.NewDense(1, 6, bowResult)
			matrixInit := mat.NewDense(1, 6, aInit)
			var sixdofA mat.Dense
			sixdofA.Sub(matrixA, matrixInit)
			input64 := sixdofA.RawRowView(0)
			input32 := make([]float32, 6)
			for i := 0; i < 6; i++ {
				input32[i] = float32(input64[i])
			}
			c.Client.Client.Execute(input32, clylen)
			if executeRequest.Period == 0 {
				time.Sleep(33 * time.Millisecond)
			} else {
				time.Sleep(time.Duration(executeRequest.Period) * time.Millisecond)
			}

			log.Printf("execute with %v", clylen)
			_, err := port.Write(c.Client.AssembleSerialData(clylen))
			if err != nil {
				log.Println("Error writing to serial port:%v ", err)
				return
			}
		}
		// reset..
		_, err = port.Write(c.Client.ResetToZero())
		if err != nil {
			log.Println("Error writing to serial port:%v ", err)
			return
		}
		// close websocket.
		service.CloseClientSocket(executeRequest.ID)
	}()
}
