package request

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"io"
	"strings"
)

const BUFFER_CAPACITY = 5
const CR_DELIMETER = '\r'

type requestState string

const (
	requestProgress requestState = "progress"
	requestDone     requestState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine *RequestLine
	Headers     *headers.Headers
	Body        []byte
}

func NewRequest() *Request {
	return &Request{Headers: headers.NewHeaders()}
}

func (r *Request) parse(line []byte) error {
	var err error
	if len(line) == 0 {
		// do nothing
	} else if r.RequestLine == nil {
		r.RequestLine, err = readRequestLine(line)
	} else {
		k, v, err := readFieldLine(line)
		if err == nil {
			r.Headers.Set(k, v)
		}
	}

	return err
}

// Read data input with dynamic buffer
//
// Outer & Inner loops are inversely proportional
func RequestFromReader(reader io.Reader) (*Request, error) {

	request := NewRequest()
	buffer := make([]byte, BUFFER_CAPACITY)

	var (
		n, startId, endId, resStartId int
		err, pErr                     error
		line, resBuffer               []byte
	)

outer:
	for {
		startId = 0
		n, err = reader.Read(buffer)
		if err != nil {
			break outer
		}

	inner:
		for i := 0; i < n; i++ {

			if startId > n {
				break inner
			}

			if buffer[i] == CR_DELIMETER {
				endId = i
			} else {
				endId = -1
				continue
			}

			// Merge residual buffer, if any
			if len(resBuffer) > 0 {
				// special case: if '\r' char is the last element of the array
				// then '\n' char, in the next loop, could be present at the first position of the array then we need to skip it
				if resBuffer[0] == '\n' {
					resStartId = 1
				}
				line = buildLine(resBuffer[resStartId:], buffer[startId:endId])
				resBuffer = resBuffer[:0]
				resStartId = 0
			} else {
				line = buildLine(nil, buffer[startId:endId])
			}

			pErr = request.parse(line)
			if pErr != nil {
				break outer
			}

			startId = endId + 2
		}

		// Accumulate residual buffer if bufferSize is very small
		if endId == -1 {
			resBuffer = append(resBuffer, buffer[startId:n]...)
		} else if endId+2 < n {
			resBuffer = append(resBuffer, buffer[endId+2:n]...)
		}
	}

	return request, pErr
}

func readRequestLine(l []byte) (*RequestLine, error) {
	splt := bytes.Split(l, []byte{' '})
	if len(splt) < 3 {
		return nil, fmt.Errorf("[Read Request Line] invalid number of parts in request line. current: %v Requested: (Method  target  http version)", splt)
	}
	requestLine := &RequestLine{}
	requestLine.Method = string(splt[0])
	requestLine.RequestTarget = string(splt[1])
	requestLine.HttpVersion = strings.TrimPrefix(string(splt[2]), "HTTP/")
	return requestLine, nil
}

func readFieldLine(l []byte) (k []byte, v []byte, e error) {
	idx := bytes.IndexByte(l, ':')
	if idx == -1 {
		return nil, nil, fmt.Errorf("it cannot possible extract key value because the ':' is missing")
	}
	key := bytes.TrimSpace(l[:idx])
	value := bytes.TrimSpace(l[idx+1:])
	return key, value, nil
}

func buildLine(a []byte, b []byte) []byte {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}

	line := make([]byte, len(a)+len(b))
	copy(line, a)
	copy(line[len(a):], b)
	return line
}
