package request

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"io"
	"strings"
)

const BUFFER_CAPACITY = 1024

// Carriage-return
const CR_DELIMETER = '\r'

// Line-feed
const LN_DELIMETER = '\n'

type RequestState string

const (
	RequestInit    RequestState = "init"
	RequestHeaders RequestState = "headers"
	RequestDone    RequestState = "done"
	RequestError   RequestState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	state       RequestState
	RequestLine *RequestLine
	Headers     *headers.Headers
	Body        []byte
}

func NewRequest() *Request {
	return &Request{state: RequestInit, Headers: headers.NewHeaders()}
}

func (r *Request) parse(line []byte) (int, error) {
	var (
		rd, read   int
		err        error
		curretLine []byte
		done       bool
	)

outer:
	for {
		curretLine = line[read:]
		switch r.state {
		case RequestError:
			break outer
		case RequestInit:
			r.RequestLine, rd, err = readRequestLine(curretLine)
			if err != nil {
				r.state = RequestError
				break outer
			}

			if rd == 0 {
				break outer
			}

			read += rd
			r.state = RequestHeaders
		case RequestHeaders:
			rd, done, err = r.Headers.ParseAll(curretLine)
			if err != nil {
				r.state = RequestError
				break outer
			}

			if rd == 0 {
				break outer
			}

			read += rd

			if done {
				r.state = RequestDone
			}

		case RequestDone:
			break outer

		default:
			panic("You are idiot")
		}
	}

	return read, err
}

// Read data input with dynamic buffer
//
// Outer & Inner loops are inversely proportional
func RequestFromReader(reader io.Reader) (*Request, error) {

	request := NewRequest()
	buffer := make([]byte, BUFFER_CAPACITY)

	var (
		n, read, startId int
		err, pErr        error
	)

outer:
	for request.state != RequestDone {
		n, err = reader.Read(buffer[startId:])
		if err != nil {
			break outer
		}

		startId += n
		// read is number of processed byte
		read, pErr = request.parse(buffer[:startId])
		if pErr != nil {
			break outer
		}

		// moves unprocessed data to the left, freeing up the buffer for new data
		if read > 0 {
			copy(buffer, buffer[read:])
			startId -= read
		}

	}

	return request, pErr
}

func readRequestLine(l []byte) (*RequestLine, int, error) {
	read := bytes.Index(l, []byte{CR_DELIMETER, LN_DELIMETER})
	if read == -1 {
		return nil, 0, nil
	}

	splt := bytes.Split(l[:read], []byte{' '})
	if len(splt) < 3 {
		return nil, 0, fmt.Errorf("invalid number of parts in request line. current: %v Requested: (Method  target  http version)", splt)
	}
	requestLine := &RequestLine{}
	requestLine.Method = string(splt[0])
	requestLine.RequestTarget = string(splt[1])
	requestLine.HttpVersion = strings.TrimPrefix(string(splt[2]), "HTTP/")
	return requestLine, read + 2, nil
}
