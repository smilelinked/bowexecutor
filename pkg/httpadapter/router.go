package httpadapter

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smilelinkd/bowexecutor/driver"
	"github.com/smilelinkd/bowexecutor/pkg/common"
	"k8s.io/klog/v2"
)

// RestController the struct of HTTP route
type RestController struct {
	Client *driver.DigitalbowClient
}

// NewRestController build a RestController
func NewRestController(dic *driver.DigitalbowClient) *RestController {
	return &RestController{
		Client: dic,
	}
}

// sendResponse puts together the response packet for the V2 API
func (c *RestController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	API string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			klog.Error(fmt.Sprintf("Unable to marshal %s response", API), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			klog.Error(fmt.Sprintf("Unable to write %s response", API), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
