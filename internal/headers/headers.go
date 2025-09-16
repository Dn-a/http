package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// Carriage-return
const CR_DELIMETER = '\r'

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

func (h *Headers) Set(k []byte, v []byte) error {
	if ok, c := isToken(k); !ok {
		return fmt.Errorf("field-value header doesn't contains a valid characters: %c", c)
	}
	h.headers[strings.ToLower(string(k))] = string(v)
	return nil
}

func (h *Headers) Parse(data []byte) (n int, done bool, er error) {

	var (
		read           int = 0
		startId, endId int
		k, v           []byte
		err            error
		check          bool = true
	)

	for i := range data {

		if data[i] == CR_DELIMETER {
			endId = i
		} else {
			continue
		}

		if k, v, err = parseHeader(data[startId:endId]); err != nil {
			check = false
			break
		}

		if err = h.Set(k, v); err != nil {
			check = false
			break
		}

		read += len(k) + len(v) + 2
		startId = endId + 2
	}
	return read, check, err
}

func parseHeader(line []byte) ([]byte, []byte, error) {
	if len(line) == 0 {
		return nil, nil, nil
	}
	splt := bytes.SplitN(line, []byte{':'}, 2)

	if len(splt) != 2 || bytes.HasSuffix(splt[0], []byte{' '}) {
		return nil, nil, fmt.Errorf("malformed Header")
	}

	return splt[0], bytes.TrimSpace(splt[1]), nil
}

// file-value VALIDATOR
var specialChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~', ':', ' ', '\r', '\n'}

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
