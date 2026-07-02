package httpclient

import (
	"net/http"
	"time"
)

const DefaultTimeout = 30 * time.Second

var Client = &http.Client{
	Timeout: DefaultTimeout,
}
