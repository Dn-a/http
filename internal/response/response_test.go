package response

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_Write(t *testing.T) {
	t.Run("should write default OK response when status and headers are nil", func(t *testing.T) {
		var buf bytes.Buffer
		res := Response{Writer: &buf}
		body := []byte("Hello0o0o0")

		res.Write(nil, nil, body)

		output := buf.String()
		expectedPrefix := "HTTP/1.1 200 OK\r\n"
		expectedBodySuffix := "\r\n\r\nHello0o0o0"

		require.True(t, strings.HasPrefix(output, expectedPrefix), "Expected status line was not found")
		assert.Contains(t, output, "content-type: text/plain\r\n")
		assert.Contains(t, output, "connection: close\r\n")
		assert.Contains(t, output, fmt.Sprintf("content-length: %s\r\n", strconv.Itoa(len(body))))
		require.True(t, strings.HasSuffix(output, expectedBodySuffix), "Body was not written correctly")
	})

	t.Run("should use provided status and headers with an empty body", func(t *testing.T) {
		var buf bytes.Buffer
		res := Response{Writer: &buf}
		status := &NOT_FOUND
		hdrs := headers.NewHeaders()
		hdrs.Set("X-Custom-Header", "Im-header")

		res.Write(status, hdrs, []byte{})

		output := buf.String()
		expectedPrefix := "HTTP/1.1 404 Not Found\r\n"
		expectedSuffix := "\r\n\r\n"

		require.True(t, strings.HasPrefix(output, expectedPrefix))
		assert.Contains(t, output, "x-custom-header: Im-header\r\n")
		assert.Contains(t, output, "content-length: 0\r\n")
		require.True(t, strings.HasSuffix(output, expectedSuffix))
	})

	t.Run("should calculate Content-Length if not provided in headers", func(t *testing.T) {
		var buf bytes.Buffer
		res := Response{Writer: &buf}
		body := []byte("Im full!")
		hdrs := headers.NewHeaders()
		hdrs.Set("X-Another-Header", "not-alone")

		res.Write(&OK, hdrs, body)

		output := buf.String()
		assert.Contains(t, output, fmt.Sprintf("content-length: %s\r\n", strconv.Itoa(len(body))))
	})
}

func TestResponse_WriteChunkedBody(t *testing.T) {
	var buf bytes.Buffer
	res := Response{Writer: &buf}

	// TEST 1
	chunkData := []byte("this is a chunk of you")
	_, err := res.WriteChunkedBody(chunkData)
	require.NoError(t, err)

	output := buf.String()
	// Length of "this is a chunk" is 22, which is '16' in hex.
	expected := "16\r\nthis is a chunk of you\r\n"
	assert.Equal(t, expected, output)

	// TEST 2: few text
	buf.Reset()
	chunkData = []byte("little chunk")
	_, err = res.WriteChunkedBody(chunkData)
	require.NoError(t, err)

	output = buf.String()
	// Length of "this is a chunk" is 12, which is 'c' in hex.
	// The function formats it as %02x, so it becomes "0c".
	expected = "0c\r\nlittle chunk\r\n"
	assert.Equal(t, expected, output)
}

func TestResponse_WriteChunkedBodyDone(t *testing.T) {
	var buf bytes.Buffer
	res := Response{Writer: &buf}

	_, err := res.WriteChunkedBodyDone()
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, "0\r\n\r\n", output)
}

func TestResponse_WriteTrailers(t *testing.T) {
	var buf bytes.Buffer
	res := Response{Writer: &buf}

	trailers := headers.NewHeaders()
	trailers.Set("X-Checksum", "abcde12345")
	trailers.Set("Expires", "Wed, 21 Oct 2025 07:28:00 GMT")

	err := res.WriteTrailers(trailers)
	require.NoError(t, err)

	output := buf.String()
	// Should start with the zero-length chunk, followed by trailers, and a final CRLF.
	expectedPrefix := "0\r\n"
	expectedSuffix := "\r\n"
	require.True(t, strings.HasPrefix(output, expectedPrefix))
	assert.Contains(t, output, "x-checksum: abcde12345\r\n")
	assert.Contains(t, output, "expires: Wed, 21 Oct 2025 07:28:00 GMT\r\n")
	require.True(t, strings.HasSuffix(output, expectedSuffix))
}
