package request

import (
	"fmt"
	"io"
	"strings"
)

const BUFFER_CAPACITY = 2

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
	Headers     map[string]string
	Body        []byte
}

func (r *Request) parse(line *string) error {
	var err error
	if r.RequestLine == nil {
		r.RequestLine, err = readRequestLine(line)
	} else {
		k, v, lineErr := readFieldLine(line)
		if lineErr == nil {
			r.Headers[k] = v
		}
	}

	return err
}

// Read data input with dynamic buffer
//
// Outer & Inner loops are inversely proportional
func RequestFromReader(reader io.Reader) (*Request, error) {

	request := &Request{Headers: make(map[string]string)}
	buffer := make([]byte, BUFFER_CAPACITY)

	var (
		n, startId, endId, resStartId int
		err, pErr                     error
		line                          *string
		resBuffer                     []byte
	)

outer:
	for {
		startId = 0
		n, err = reader.Read(buffer)
		if err != nil {
			break outer
		}

	inner:
		for {

			endId = strings.IndexByte(string(buffer[startId:n]), '\r')
			if endId == -1 {
				break inner
			}
			endId += startId

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

			if line == nil {
				break inner
			}

			pErr = request.parse(line)
			if pErr != nil {
				break outer
			}

			if endId+2 > len(buffer) {
				break inner
			}

			startId = endId + 2
		}

		// Accumulate residual buffer if bufferSize is very small
		if endId == -1 {
			resBuffer = append(resBuffer, buffer[startId:]...)
		} else if endId+2 < len(buffer) {
			resBuffer = append(resBuffer, buffer[endId+2:]...)
		}
	}

	return request, pErr
}

func readRequestLine(l *string) (*RequestLine, error) {
	splt := strings.Split(*l, " ")
	if len(splt) < 3 {
		return nil, fmt.Errorf("[Read Request Line] invalid number of parts in request line. current: %v Requested: (Method  target  http version)", splt)
	}
	requestLine := &RequestLine{}
	requestLine.Method = splt[0]
	requestLine.RequestTarget = splt[1]
	requestLine.HttpVersion = strings.TrimPrefix(splt[2], "HTTP/")
	return requestLine, nil
}

func readFieldLine(l *string) (k string, v string, e error) {
	idx := strings.IndexByte(*l, ':')
	if idx == -1 {
		return "", "", fmt.Errorf("It cannot possible extract key value because the ':' is missing")
	}
	key := strings.TrimSpace((*l)[:idx])
	value := strings.TrimSpace((*l)[idx+1:])
	return key, value, nil
}

func buildLine(a []byte, b []byte) *string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.Grow(len(a) + len(b))
	if a != nil {
		builder.Write(a)
	}
	if b != nil {
		builder.Write(b)
	}
	str := builder.String()
	return &str
}
