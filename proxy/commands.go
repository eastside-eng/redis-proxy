package proxy

import (
	"errors"
	"strconv"
	"strings"
)

// Command is a parsed version of the Redis RESP protocol. Each Command
// should implement a Handler function that takes a RedisClient, Cache and processes
// the request.
type Command struct {
	Name string
	Args []string
}

// Scans the byte array until it hits a CLRF.
// Returns the string from [start, index of CLRF] and the next string's start.
func nextString(raw []byte, start int) (string, int, error) {
	stop := start
	for raw[stop] != '\r' && stop < len(raw) {
		stop++
	}
	if stop > len(raw) {
		return "", 0, errors.New("Invalid input given to nextString")
	}
	return string(raw[start:stop]), stop + 2, nil
}

// ParseCommand takes a byte array and parses it into a Command instance.
// e.g. *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\nv -> ["FOO" "BAR"]
func parseCommand(raw []byte) (*Command, error) {
	// parse number of items
	str, stop, err := nextString(raw, 1)
	if err != nil {
		return nil, err
	}

	numItems, err := strconv.Atoi(str)
	if err != nil {
		return nil, err
	}

	res := make([]string, numItems)
	cnt := 0
	for cnt = 0; cnt < numItems && stop < len(raw); cnt++ {
		// We parse the next item's length. It starts with a '$', so we skip 1.
		str, start, err := nextString(raw, stop+1)
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		// Parse the next length bytes directly into res.
		res[cnt] = string(raw[start : start+length])
		stop = start + length + 2 // move to next string
	}

	if cnt != len(res) || cnt == 0 {
		return nil, errors.New("Invalid byte array given to parseCommand")
	}

	if cnt == 1 {
		return &Command{Name: strings.ToUpper(res[0]), Args: make([]string, 0)}, nil
	}

	return &Command{Name: strings.ToUpper(res[0]), Args: res[1:]}, nil
}
