package pkg

import (
	"net/http"
	"time"

	"github.com/smilelinkd/bowexecutor/pkg/httpadapter"
)

// HTTPClient is structure used to init HttpClient
type HTTPClient struct {
	IP             string
	Port           string
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	server         *http.Server
	restController *httpadapter.RestController
}
