package format

import (
	"bufio"
	"bytes"
	"strconv"

	"github.com/tea4go/gh/syslog/internal/syslogparser/rfc5424"
)

type RFC6587 struct{}

func (f *RFC6587) GetParser(line []byte) LogParser {
	return &parserWrapper{rfc5424.NewParser(line)}
}

func (f *RFC6587) GetSplitFunc() bufio.SplitFunc {
	return rfc6587ScannerSplit
}

func rfc6587ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, ' '); i > 0 {
		pLength := data[0:i]
		length, err := strconv.Atoi(string(pLength))
		if err != nil {
			if string(data[0:1]) == "<" {
				// Assume this frame uses non-transparent-framing
				return len(data), data, nil
			}
			return 0, nil, err
		}
		end := length + i + 1
		// 防止恶意/畸形长度导致整型溢出或负值，进而 data[i+1:end] 越界 panic
		if length < 0 || end < i+1 {
			return 0, nil, strconv.ErrRange
		}
		if len(data) >= end {
			// Return the frame with the length removed
			return end, data[i+1 : end], nil
		}
	}

	// Request more data
	return 0, nil, nil
}
