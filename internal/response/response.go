package response

import (
	"fmt"
	"http/internal/headers"
	"io"
	"strconv"
)

type StatusCode struct {
	Reason string
	Code   uint16
}
type Response struct {
	Writer io.Writer
}

var (
	OK                    StatusCode = StatusCode{"OK", 200}
	NOT_FOUND             StatusCode = StatusCode{"Not Found", 404}
	BAD_REQUEST           StatusCode = StatusCode{"Bad Request", 400}
	INTERNAL_SERVER_ERROR StatusCode = StatusCode{"Internal Server Error", 500}
)

const HTTP_VERSION = "HTTP/1.1"

const DELIMITER = "\r\n"

func (res *Response) Write(status *StatusCode, currentHeaders *headers.Headers, body []byte) {
	if status == nil {
		status = &OK
	}
	writeStatusLine(res.Writer, status)

	if currentHeaders == nil {
		currentHeaders = GetDefaultHeaders(len(body))
	} else if currentHeaders.GetContentLength() == 0 && len(body) > 0 {
		currentHeaders.Set(headers.CONTENT_LENGTH, strconv.Itoa(len(body)))
	}
	writeHeaders(res.Writer, currentHeaders)

	if len(body) > 0 {
		writeBody(res.Writer, body)
	}
}

func (r *Response) WriteChunkedBody(size []byte, p []byte) (int, error) {
	r.Writer.Write(size)
	return r.Writer.Write(append(p, []byte(DELIMITER)...))
}
func (r *Response) WriteChunkedBodyDone() (int, error) {
	r.Writer.Write([]byte{'0', '\r', '\n', '\r', '\n'})
	return 0, nil
}

func writeStatusLine(w io.Writer, statusCode *StatusCode) error {
	fmt.Fprintf(w, "%v %v %v\r\n", HTTP_VERSION, statusCode.Code, statusCode.Reason)
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
	h.Set(headers.CONTENT_TYPE, "text/plain")
	h.Set(headers.CONTENT_LENGTH, strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	return h
}
