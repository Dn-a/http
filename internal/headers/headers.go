package headers

import (
	"bytes"
	"fmt"
)

const CR_DELIMETER = '\r'

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, er error) {

	var (
		read           int = 0
		startId, endId int
		k, v           string
		err            error
	)

	for i := 0; i < len(data); i++ {
		if data[i] == CR_DELIMETER {
			endId = i
		} else {
			continue
		}

		k, v, err = parseHeader(data[startId:endId])
		if err != nil {
			return 0, false, err
		}
		h[k] = v
		read += len(k) + len(v) + 2
		startId = endId + 2
	}
	return read, false, nil
}

func parseHeader(line []byte) (string, string, error) {
	if len(line) == 0 {
		return "", "", nil
	}
	splt := bytes.SplitN(line, []byte{':'}, 2)

	if len(splt) != 2 || bytes.HasSuffix(splt[0], []byte{' '}) {
		return "", "", fmt.Errorf("Malformed Header.")
	}

	key := string(splt[0])
	value := string(bytes.TrimSpace(splt[1]))

	return key, value, nil
}
