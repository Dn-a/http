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
type Response struct {
	headers *headers.Headers
	body    []byte
	Writer  io.Writer
	status  *StatusCode
}

var (
	OK                    StatusCode = StatusCode{"OK", 200}
	NOT_FOUND             StatusCode = StatusCode{"Not Found", 404}
	BAD_REQUEST           StatusCode = StatusCode{"Bad Request", 400}
	INTERNAL_SERVER_ERROR StatusCode = StatusCode{"Internal Server Error", 500}
)

const HTTP_VERSION = "HTTP/1.1"

func (res *Response) Status(s *StatusCode) {
	res.status = s
}

func (res *Response) Headers(h *headers.Headers) {
	res.headers = h
}
func (res *Response) Body(b []byte) {
	res.body = append(b, '\n')
}

func (res *Response) Write() {
	if res.status == nil {
		res.status = &OK
	}
	writeStatusLine(res.Writer, *res.status)
	if res.headers == nil {
		res.headers = GetDefaultHeaders(len(res.body))
	}
	writeHeaders(res.Writer, res.headers)
	if len(res.body) > 0 {
		writeBody(res.Writer, res.body)
	}
}

func writeStatusLine(w io.Writer, statusCode StatusCode) error {
	fmt.Fprintf(w, "%v %v %v\r\n", HTTP_VERSION, statusCode.code, statusCode.reason)
	return nil
}

func writeHeaders(w io.Writer, headers *headers.Headers) error {
	headers.ForEach(func(k, v string) {
		fmt.Fprintf(w, "%v: %v\r\n", k, v)
	})
	fmt.Fprint(w, "\r\n")
	return nil
}

func writeBody(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Type", "text/plain")
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	return h
}
