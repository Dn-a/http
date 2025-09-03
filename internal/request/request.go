package request

import (
	"fmt"
	"io"
	"strings"
)

const BUFFER_CAPACITY = 4

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

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := &Request{Headers: make(map[string]string)}

	buffer := make([]byte, BUFFER_CAPACITY)
	resBuffer := make([]byte, 0, BUFFER_CAPACITY)

	var (
		endId   int
		startId int
		//n           int
		err         error
		gError      error
		isFirstLine bool = true
		k, v        string
		line        *string
	)

	for {
		_, err = reader.Read(buffer)
		if err != nil || gError != nil {
			break
		}

		for {

			endId = strings.IndexByte(string(buffer[:len(buffer)]), '\r')

			if endId != -1 {

				if endId > len(buffer) {
					break
				}

				// Merge residual buffer, if any
				if len(resBuffer) > 0 {
					// special case: if '\r' char is the last element of the array
					// then '\n' char, in the next loop, could be present at the first position of the array then we need to skip it
					if resBuffer[0] == '\n' {
						startId = 1
					}
					line = buildLine(resBuffer[startId:], buffer[:endId])
					resBuffer = resBuffer[:0]
					startId = 0
				} else {
					line = buildLine(nil, buffer[:endId])
				}

				if line == nil {
					break
				}

				// Parser
				if isFirstLine {
					request.RequestLine, gError = readRequestLine(line)
					isFirstLine = false
				} else if !isFirstLine {
					k, v = readFieldLine(line)
					request.Headers[k] = v
				}

				if gError != nil {
					break
				}

				if endId+2 < len(buffer) {
					//n = len(buffer[endId+2 : n])
					buffer = buffer[endId+2:]
				} else {
					break
				}

			} else {
				break
			}

		}

		if endId == -1 {
			resBuffer = append(resBuffer, buffer[:len(buffer)]...)
		} else if endId+2 < len(buffer) {
			resBuffer = append(resBuffer, buffer[endId+2:len(buffer)]...)
		}
		//buffer = buffer[:0]
	}

	return request, gError
}

func readRequestLine(l *string) (*RequestLine, error) {
	requestLine := &RequestLine{}
	splt := strings.Split(*l, " ")
	if len(splt) < 3 {
		return nil, fmt.Errorf("[Read Request Line] invalid number of parts in request line. current: %v Requested: (Method  target  http version)", splt)
	}
	requestLine.Method = splt[0]
	requestLine.RequestTarget = splt[1]
	requestLine.HttpVersion = strings.TrimPrefix(splt[2], "HTTP/")
	return requestLine, nil
}

func readFieldLine(l *string) (k string, v string) {
	idx := strings.IndexByte(*l, ':')
	key := strings.TrimSpace((*l)[:idx])
	value := strings.TrimSpace((*l)[idx+1:])
	return key, value
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
