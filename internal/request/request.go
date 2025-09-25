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
	RequestBody    RequestState = "body"
	RequestDone    RequestState = "done"
	RequestError   RequestState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	Body        []byte
	Headers     *headers.Headers
	state       RequestState
	RequestLine *RequestLine
	bodyRead    int
	isEof       bool
}

func NewRequest() *Request {
	return &Request{state: RequestInit, Headers: headers.NewHeaders()}
}

func (r *Request) parse(line []byte) (int, error) {
	var (
		curretLine               []byte
		err                      error
		rd, read, curretLineSize int
		done                     bool
	)

outer:
	for {
		curretLine = line[read:]
		curretLineSize = len(curretLine)
		switch r.state {
		case RequestError:
			return 0, fmt.Errorf("general error during parsing request")
		case RequestInit:
			r.RequestLine, rd, err = readRequestLine(curretLine)
			if err != nil {
				r.state = RequestError
				break outer
			}

			// when rd == 0, there isn't enough data in the buffer to build the requestLine
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

			// If parsing reaches the last field-line, we can assume there are no other field-lines
			// e.g.: accept: */*\r\n\r\n -> The double CRLF (\r\n\r\n) is the proper delimiter
			// between HTTP headers and message body according to RFC 7230
			if done {
				if r.Headers.GetContentLength() == 0 {
					r.state = RequestDone
				} else {
					r.state = RequestBody
				}
			} else if rd == 0 {
				// when rd == 0, there isn't enough data in the buffer to build the header
				break outer
			}
			read += rd
		case RequestBody:
			length := r.Headers.GetContentLength()

			if r.Body == nil {
				r.Body = make([]byte, length)
			}
			copy(r.Body[r.bodyRead:], curretLine)

			r.bodyRead += curretLineSize
			read += curretLineSize

			if r.bodyRead > length || (r.isEof && r.bodyRead < length) {
				err = fmt.Errorf("body cannot be shorter or greater then Content-length.\n - content-length: %v\n - bodyRead: %v", length, r.bodyRead)
				r.state = RequestError
				break outer
			} else if r.bodyRead == length || r.isEof {
				r.state = RequestDone
			} else if read >= curretLineSize {
				break outer
			}
		case RequestDone:
			fmt.Println("All data are consumed correctly")
			break outer
		default:
			panic("You are idiot")
		}
	}

	return read, err
}

// Read data input with dynamic buffer
func RequestFromReader(reader io.Reader) (*Request, error) {

	request := NewRequest()
	buffer := make([]byte, BUFFER_CAPACITY)

	var (
		err, pErr        error
		n, read, startId int
	)

outer:
	for request.state != RequestDone {

		n, err = reader.Read(buffer[startId:])
		if err != nil {
			request.isEof = true
		}

		startId += n

		// read is number of processed byte
		// 	- read <= startId
		read, pErr = request.parse(buffer[:startId])
		if pErr != nil {
			break outer
		}

		// moves unprocessed data to the left, freeing up the buffer for new data
		if read > 0 {
			copy(buffer, buffer[read:startId])
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

func (r *Request) PrintRequest() {
	fmt.Println("Request Line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	r.Headers.ForEach(func(k, v string) {
		fmt.Printf("- %s: %s\n", k, v)
	})
	fmt.Printf("Body:\n%s\n", string(r.Body))
}
