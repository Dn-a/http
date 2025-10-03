package response

import (
	"fmt"
	"http/components/headers"
	"io"
	"strconv"
	"strings"
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
	} else if body != nil {
		currentHeaders.Set(headers.CONTENT_LENGTH, strconv.Itoa(len(body)))
	}
	writeHeaders(res.Writer, currentHeaders)

	if len(body) > 0 {
		writeBody(res.Writer, body)
	}
}

// TODO: need refactor
func (r *Response) WriteChunkedBody(p []byte) (int, error) {
	_, err := r.Writer.Write(fmt.Appendf(nil, "%02x\r\n", len(p))) // write chunk size
	if err != nil {
		return 0, err
	}
	_, err = r.Writer.Write(p) // write chunk
	if err != nil {
		return 0, err
	}
	_, err = r.Writer.Write([]byte(DELIMITER)) // delimiter
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func (r *Response) WriteTrailers(h *headers.Headers) error {
	var hBuilder strings.Builder

	// Signal end of the body
	hBuilder.Write([]byte{'0', '\r', '\n'})

	h.ForEach(func(k, v string) {
		hBuilder.WriteString(k)
		hBuilder.Write([]byte{':', ' '})
		hBuilder.WriteString(v)
		hBuilder.Write([]byte{'\r', '\n'})
	})

	// signal end of the trailer
	hBuilder.Write([]byte{'\r', '\n'})

	_, error := io.WriteString(r.Writer, hBuilder.String())
	return error
}

func (r *Response) WriteChunkedBodyDone() (int, error) {
	r.Writer.Write([]byte{'0', '\r', '\n', '\r', '\n'})
	return 0, nil
}

func writeStatusLine(w io.Writer, statusCode *StatusCode) error {
	fmt.Fprintf(w, "%v %v %v\r\n", HTTP_VERSION, statusCode.Code, statusCode.Reason)
	return nil
}

func writeHeaders(w io.Writer, h *headers.Headers) (int, error) {
	var hBuilder strings.Builder
	h.ForEach(func(k, v string) {
		hBuilder.WriteString(k)
		hBuilder.Write([]byte{':', ' '})
		hBuilder.WriteString(v)
		hBuilder.Write([]byte{'\r', '\n'})
	})
	hBuilder.Write([]byte{'\r', '\n'})

	return io.WriteString(w, hBuilder.String())
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
