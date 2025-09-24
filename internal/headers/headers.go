package headers

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Carriage-return
const CR_DELIMETER = '\r'

// Line-feed
const LN_DELIMETER = '\n'

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(v string) string {
	return h.headers[strings.ToLower(v)]
}

func (h *Headers) GetContentLength() int {
	len := h.Get("Content-Length")
	if len == "" {
		return 0
	}
	val, _ := strconv.Atoi(len)
	return val
}

func (h *Headers) Set(k string, v string) bool {
	if k != "" && v != "" {
		h.headers[strings.ToLower(k)] = v
		return true
	}
	return false
}

// Parse bytes that should contains valid field-value and line-separator (\r\n)
func (h *Headers) ParseAll(data []byte) (read int, done bool, er error) {

	var (
		startId, endId int
		k, v           []byte
		err            error
		check          bool
	)

	for i := range data {

		if data[i] == CR_DELIMETER && data[(i+1)%len(data)] == LN_DELIMETER {
			endId = i
		} else {
			continue
		}

		if k, v, err = parseHeader(data[startId:endId]); err == nil {
			if k != nil {
				h.Set(string(k), string(v))
			} else {
				// HEADER is EMPTY, so we assume there are no more headers to parse
				check = true
			}
		} else {
			break
		}
		startId = endId + 2
	}
	return startId, check, err
}

// Parse single header without \r\n
func (h *Headers) Parse(data []byte) (read int, er error) {

	var (
		rd   int = 0
		k, v []byte
		err  error
	)

	if k, v, err = parseHeader(data); err == nil {
		h.Set(string(k), string(v))
	}

	return rd, err
}

func (h *Headers) ForEach(cb func(k, v string)) {
	for k, v := range h.headers {
		cb(k, v)
	}
}

func parseHeader(line []byte) ([]byte, []byte, error) {
	if len(line) == 0 {
		return nil, nil, nil
	}
	splt := bytes.SplitN(line, []byte{':'}, 2)

	if len(splt) != 2 || bytes.HasSuffix(splt[0], []byte{' '}) {
		return nil, nil, fmt.Errorf("malformed Header")
	}

	k, v := bytes.TrimSpace(splt[0]), bytes.TrimSpace(splt[1])

	if ok, c := isToken(k); !ok {
		return nil, nil, fmt.Errorf("field-value header doesn't contains a valid characters: %c", c)
	}

	return k, v, nil
}

// field-value VALIDATOR
var specialChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func isToken(chars []byte) (bool, byte) {
	for _, c := range chars {
		if !(bytes.Contains(specialChars, []byte{c}) ||
			'a' <= c && c <= 'z' ||
			'A' <= c && c <= 'Z' ||
			'0' <= c && c <= '9') {
			return false, c
		}
	}
	return true, 0
}
