package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dream11/livelogs/pkg/logger"
)

// Request structure
type Request struct {
	Method  string
	URL     string
	Header  map[string]string
	Query   map[string]string
	Body    interface{}
	Timeout time.Duration
}

// Response structure
type Response struct {
	Status     string
	StatusCode int
	Body       []byte
	Error      error
}

const defaultRequestTimeOut = 5 * time.Second

var log logger.Logger

// Make : make a generated request
func (r *Request) Make() Response {
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(r.Body)
	if err != nil {
		return Response{Error: err}
	}

	request, err := http.NewRequest(r.Method, r.URL, payload)
	if err != nil {
		return Response{Error: err}
	}

	q := request.URL.Query()
	for key, val := range r.Query {
		if len(val) > 0 {
			q.Add(key, val)
		}
	}
	request.URL.RawQuery = q.Encode()

	log.Debug("URL: " + request.URL.String())

	for key, value := range r.Header {
		if len(value) > 0 {
			request.Header.Set(key, value)
		}
	}

	if r.Timeout == 0 {
		r.Timeout = defaultRequestTimeOut
	}
	response, err := (&http.Client{Timeout: r.Timeout}).Do(request)

	if err != nil {
		return Response{Error: err}
	}

	respBody, err := io.ReadAll(response.Body)

	if err != nil {
		return Response{Error: err}
	}

	return Response{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Body:       respBody,
		Error:      nil,
	}
}
