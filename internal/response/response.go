package response

import (
	"fmt"
	"http/internal/headers"
	"io"
	"strconv"
)

type StatusCode struct {
	reason string
	code   uint16
}

var (
	OK                    StatusCode = StatusCode{"OK", 200}
	BAD_REQUEST           StatusCode = StatusCode{"Bad Request", 400}
	INTERNAL_SERVER_ERROR StatusCode = StatusCode{"Internal Server Error", 500}
)

const HTTP_VERSION = "HTTP/1.1"

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	fmt.Fprintf(w, "%v %v %v\r\n", HTTP_VERSION, statusCode.code, statusCode.reason)
	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	headers.ForEach(func(k, v string) {
		fmt.Fprintf(w, "%v: %v\r\n", k, v)
	})
	return nil
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Type", "text/plain")
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	return h
}
