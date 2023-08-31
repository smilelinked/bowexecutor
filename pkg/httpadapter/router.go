package httpadapter

import (
	"github.com/smilelinkd/bowexecutor/driver"
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
