package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:    localhost:42069      \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: 2 headers
	headers = NewHeaders()

	// First call: Parse the Host header
	data = []byte("Host: localhost:42069\r\nAuth: Bearer token\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err, "First parse call should not return an error")
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"], "The Host header value is incorrect")
	assert.False(t, done, "Done should still be false because there are more headers to process")
	assert.Equal(t, 23, n, "Should only consume bytes for the first header ('Host')")

	// Second call: Parse the Auth header
	remainingData := data[n:] // Use the remaining unparsed bytes
	n, done, err = headers.Parse(remainingData)
	require.NoError(t, err, "Second parse call should not return an error")
	require.NotNil(t, headers)
	assert.Equal(t, "Bearer token", headers["auth"], "The Auth header value is incorrect")
	assert.False(
		t,
		done,
		"Done should still be false till parse is called again with the remaining data",
	)
	assert.Equal(t, 20, n, "Should consume bytes for the second header ('Auth')")

	remainingData = remainingData[n:]
	n, done, err = headers.Parse(remainingData)
	require.NoError(t, err, "Last call to parse should NOT return an error")
	require.NotNil(t, headers)
	assert.True(t, done, "Done should now be true as all header data should be parsed")

	// Test: Invalid field name
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\nAuth: Bearer token\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err, "Parse should result in invalid field name value")
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test same header with multiple values
	headers = NewHeaders()
	data = []byte(
		"Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: moose-loves-rust\r\n\r\n",
	)
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go", headers["set-person"])
	assert.Equal(t, 27, n)
	assert.False(t, done)

	remainingData = data[n:]
	n, done, err = headers.Parse(remainingData)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.Equal(t, 29, n)
	assert.False(t, done)

	remainingData = remainingData[n:]
	n, done, err = headers.Parse(remainingData)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, moose-loves-rust", headers["set-person"])
	assert.Equal(t, 30, n)
	assert.False(t, done)
}
