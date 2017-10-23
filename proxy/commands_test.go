package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	raw := []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	out, err := parseCommand(raw)
	assert.Nil(t, err)
	assert.Equal(t, "FOO", out.Name)
	assert.Equal(t, "bar", out.Args[0])
}

func TestParserFailure(t *testing.T) {
	raw := []byte("*zzz\r\n$xxx\r\nfoo\r\n$3\r\nbar\r\n")
	out, err := parseCommand(raw)
	assert.NotNil(t, err)
	assert.Nil(t, out)

	raw = []byte("*123\r\n")
	out, err = parseCommand(raw)
	assert.NotNil(t, err)
	assert.Nil(t, out)
}

func TestParserSetCommand(t *testing.T) {
	raw := []byte("*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$8\r\nmy value\r\n")
	out, _ := parseCommand(raw)
	assert.Equal(t, "SET", out.Name)
	assert.Equal(t, "mykey", out.Args[0])
	assert.Equal(t, "my value", out.Args[1])
}

func TestParserPingCommand(t *testing.T) {
	raw := []byte("*1\r\n$4\r\nPING\r\n")
	out, err := parseCommand(raw)
	assert.Nil(t, err)
	assert.Equal(t, "PING", out.Name)
	assert.Equal(t, 0, len(out.Args))
}
