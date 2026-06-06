package format

import (
	"bufio"
	"bytes"
	"strconv"
	"testing"
)

// --- RFC6587 GetParser ---

func TestRFC6587_GetParser(t *testing.T) {
	f := &RFC6587{}
	line := []byte("<34>1 2003-10-11T22:14:15Z mymachine su - ID47 - test")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := parser.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
}

// --- RFC6587 GetSplitFunc ---

func TestRFC6587_GetSplitFunc(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()
	if sf == nil {
		t.Fatal("GetSplitFunc returned nil for RFC6587")
	}
}

// --- RFC6587 splitter: well-formed frame ---

func TestRFC6587_Splitter_WellFormed(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// "24 <34>1 2003-10-11T22:14:15Z" — length prefix is 24
	payload := "<34>1 2003-10-11T22:14:15Z"
	data := []byte(strconv.Itoa(len(payload)) + " " + payload)

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != len(data) {
		t.Errorf("advance = %d, want %d", advance, len(data))
	}
	if string(token) != payload {
		t.Errorf("token = %q, want %q", string(token), payload)
	}
}

// --- RFC6587 splitter: empty input with atEOF ---

func TestRFC6587_Splitter_EmptyAtEOF(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	advance, token, err := sf([]byte{}, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 {
		t.Errorf("advance = %d, want 0", advance)
	}
	if token != nil {
		t.Errorf("token = %v, want nil", token)
	}
}

// --- RFC6587 splitter: empty input without atEOF ---

func TestRFC6587_Splitter_EmptyNotEOF(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	advance, token, err := sf([]byte{}, false)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 {
		t.Errorf("advance = %d, want 0", advance)
	}
	if token != nil {
		t.Errorf("token = %v, want nil", token)
	}
}

// --- RFC6587 splitter: partial data (need more) ---

func TestRFC6587_Splitter_PartialData(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Length prefix says 100 bytes but only 5 available
	data := []byte("100 <34>")
	advance, token, err := sf(data, false)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 {
		t.Errorf("advance = %d, want 0 (request more data)", advance)
	}
	if token != nil {
		t.Errorf("token = %v, want nil", token)
	}
}

// --- RFC6587 splitter: non-transparent framing (starts with '<') ---

func TestRFC6587_Splitter_NonTransparentFraming(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Data starting with '<' — non-transparent framing
	data := []byte("<34>Jan 11 22:14:15 mymachine su: test")
	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != len(data) {
		t.Errorf("advance = %d, want %d", advance, len(data))
	}
	if !bytes.Equal(token, data) {
		t.Errorf("token = %q, want %q", string(token), string(data))
	}
}

// --- RFC6587 splitter: malformed length (not a number, not '<') ---

func TestRFC6587_Splitter_MalformedLength(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// "abc <34>" — 'abc' is not a valid length and doesn't start with '<'
	data := []byte("abc <34>")
	_, _, err := sf(data, true)
	if err == nil {
		t.Error("expected error for malformed length")
	}
}

// --- RFC6587 splitter: integer overflow (malicious length) ---

func TestRFC6587_Splitter_IntegerOverflow(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Very large length that would cause overflow
	data := []byte("99999999999999999999 <34>")
	_, _, err := sf(data, true)
	if err == nil {
		t.Error("expected error for integer overflow in length")
	}
}

// --- RFC6587 splitter: negative-length result via overflow ---

func TestRFC6587_Splitter_NegativeLengthOverflow(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Craft a length that when added to the space position overflows
	// Length = MaxInt, i = 1 (single digit length prefix), end = MaxInt + 2 = overflow
	bigLen := strconv.Itoa(int(^uint(0) >> 1)) // MaxInt
	data := []byte(bigLen + " <34>")
	_, _, err := sf(data, true)
	// Should return ErrRange due to overflow protection
	if err != strconv.ErrRange {
		t.Errorf("expected strconv.ErrRange, got %v", err)
	}
}

// --- RFC6587 splitter: two frames in one buffer ---

func TestRFC6587_Splitter_TwoFrames(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	frame1 := "<34>1 2003-10-11T22:14:15Z host1 app - - - msg1"
	frame2 := "<165>1 2003-10-11T22:14:16Z host2 app2 - - - msg2"
	data := []byte(strconv.Itoa(len(frame1)) + " " + frame1 + strconv.Itoa(len(frame2)) + " " + frame2)

	// First call should return first frame
	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if string(token) != frame1 {
		t.Errorf("first token = %q, want %q", string(token), frame1)
	}

	// Second call with remaining data
	remaining := data[advance:]
	advance2, token2, err := sf(remaining, true)
	if err != nil {
		t.Fatalf("splitter error on second frame: %v", err)
	}
	if string(token2) != frame2 {
		t.Errorf("second token = %q, want %q", string(token2), frame2)
	}
	if advance2 != len(remaining) {
		t.Errorf("second advance = %d, want %d", advance2, len(remaining))
	}
}

// --- RFC6587 splitter: zero-length frame ---

func TestRFC6587_Splitter_ZeroLengthFrame(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	data := []byte("0 ")
	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 2 {
		t.Errorf("advance = %d, want 2", advance)
	}
	if len(token) != 0 {
		t.Errorf("token = %q, want empty", string(token))
	}
}

// --- RFC6587 splitter: length with no space yet (no space found) ---

func TestRFC6587_Splitter_NoSpaceYet(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Just digits, no space yet
	data := []byte("123")
	advance, token, err := sf(data, false)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 {
		t.Errorf("advance = %d, want 0", advance)
	}
	if token != nil {
		t.Errorf("token = %v, want nil", token)
	}
}

// --- RFC5424 GetParser ---

func TestRFC5424_GetParser(t *testing.T) {
	f := &RFC5424{}
	line := []byte("<34>1 2003-10-11T22:14:15Z mymachine su - ID47 - test")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := parser.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
	if parts["version"] != 1 {
		t.Errorf("version = %v, want 1", parts["version"])
	}
}

// --- RFC5424 GetSplitFunc (should return nil) ---

func TestRFC5424_GetSplitFunc(t *testing.T) {
	f := &RFC5424{}
	sf := f.GetSplitFunc()
	if sf != nil {
		t.Errorf("GetSplitFunc should return nil for RFC5424, got %v", sf)
	}
}

// --- RFC3164 GetParser ---

func TestRFC3164_GetParser(t *testing.T) {
	f := &RFC3164{}
	line := []byte("<34>Oct 11 22:14:15 myhost su: test message")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := parser.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
	if parts["hostname"] != "myhost" {
		t.Errorf("hostname = %v, want myhost", parts["hostname"])
	}
}

// --- RFC3164 GetSplitFunc (should return nil) ---

func TestRFC3164_GetSplitFunc(t *testing.T) {
	f := &RFC3164{}
	sf := f.GetSplitFunc()
	if sf != nil {
		t.Errorf("GetSplitFunc should return nil for RFC3164, got %v", sf)
	}
}

// --- Automatic GetParser: detect RFC3164 ---

func TestAutomatic_GetParser_RFC3164(t *testing.T) {
	f := &Automatic{}
	line := []byte("<34>Oct 11 22:14:15 myhost su: test message")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := parser.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
}

// --- Automatic GetParser: detect RFC5424 ---

func TestAutomatic_GetParser_RFC5424(t *testing.T) {
	f := &Automatic{}
	line := []byte("<34>1 2003-10-11T22:14:15Z mymachine su - ID47 - test")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := parser.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
	if parts["version"] != 1 {
		t.Errorf("version = %v, want 1", parts["version"])
	}
}

// --- Automatic GetParser: detect RFC6587 (falls back to RFC3164) ---

func TestAutomatic_GetParser_RFC6587(t *testing.T) {
	f := &Automatic{}
	// RFC6587 line with length prefix — detect will identify it as RFC6587,
	// but GetParser falls back to RFC3164 parser
	line := []byte("24 <34>1 2003-10-11T22:14:15Z")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil")
	}
}

// --- Automatic GetParser: unknown format (falls back to RFC3164) ---

func TestAutomatic_GetParser_UnknownFormat(t *testing.T) {
	f := &Automatic{}
	// No space at all — detect returns detectedUnknown
	line := []byte("nospacemessage")
	parser := f.GetParser(line)
	if parser == nil {
		t.Fatal("GetParser returned nil for unknown format")
	}
}

// --- Automatic GetSplitFunc ---

func TestAutomatic_GetSplitFunc(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()
	if sf == nil {
		t.Fatal("GetSplitFunc returned nil for Automatic")
	}
}

// --- Automatic splitter: RFC6587 frame ---

func TestAutomatic_Splitter_RFC6587Frame(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	payload := "<34>1 2003-10-11T22:14:15Z host app - - - test"
	data := []byte(strconv.Itoa(len(payload)) + " " + payload)

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != len(data) {
		t.Errorf("advance = %d, want %d", advance, len(data))
	}
	if string(token) != payload {
		t.Errorf("token = %q, want %q", string(token), payload)
	}
}

// --- Automatic splitter: RFC3164 frame (line-based) ---

func TestAutomatic_Splitter_RFC3164Frame(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	data := []byte("<34>Oct 11 22:14:15 myhost su: test message\n")

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance == 0 {
		t.Error("advance = 0, expected data to be consumed")
	}
	if token == nil {
		t.Error("token is nil, expected data")
	}
}

// --- Automatic splitter: RFC5424 frame (line-based) ---

func TestAutomatic_Splitter_RFC5424Frame(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	data := []byte("<34>1 2003-10-11T22:14:15Z mymachine su - ID47 - test\n")

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance == 0 {
		t.Error("advance = 0, expected data to be consumed")
	}
	if token == nil {
		t.Error("token is nil, expected data")
	}
}

// --- Automatic splitter: empty at EOF ---

func TestAutomatic_Splitter_EmptyAtEOF(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	advance, token, err := sf([]byte{}, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 || token != nil {
		t.Errorf("expected (0, nil), got (%d, %v)", advance, token)
	}
}

// --- Automatic splitter: no space (unknown format, request more data) ---

func TestAutomatic_Splitter_NoSpace(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	data := []byte("nospacemessage")
	advance, token, err := sf(data, false)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != 0 {
		t.Errorf("advance = %d, want 0 (request more data)", advance)
	}
	if token != nil {
		t.Errorf("token = %v, want nil", token)
	}
}

// --- detect function tests ---

func TestDetect_RFC6587(t *testing.T) {
	data := []byte("24 <34>1 2003-10-11T22:14:15Z")
	format, err := detect(data)
	if err != nil {
		t.Fatalf("detect error: %v", err)
	}
	if format != detectedRFC6587 {
		t.Errorf("format = %d, want detectedRFC6587 (%d)", format, detectedRFC6587)
	}
}

func TestDetect_RFC5424(t *testing.T) {
	data := []byte("<34>1 2003-10-11T22:14:15Z host app - - - test")
	format, err := detect(data)
	if err != nil {
		t.Fatalf("detect error: %v", err)
	}
	if format != detectedRFC5424 {
		t.Errorf("format = %d, want detectedRFC5424 (%d)", format, detectedRFC5424)
	}
}

func TestDetect_RFC3164(t *testing.T) {
	data := []byte("<34>Oct 11 22:14:15 myhost su: test")
	format, err := detect(data)
	if err != nil {
		t.Fatalf("detect error: %v", err)
	}
	if format != detectedRFC3164 {
		t.Errorf("format = %d, want detectedRFC3164 (%d)", format, detectedRFC3164)
	}
}

func TestDetect_Unknown(t *testing.T) {
	data := []byte("nospacemessage")
	format, err := detect(data)
	if err != nil {
		t.Fatalf("detect error: %v", err)
	}
	if format != detectedUnknown {
		t.Errorf("format = %d, want detectedUnknown (%d)", format, detectedUnknown)
	}
}

func TestDetect_NoCloseAngleBracket(t *testing.T) {
	// Space before close angle bracket
	data := []byte("noangle test message")
	format, err := detect(data)
	if err == nil {
		t.Errorf("expected error for no close angle bracket, got format=%d", format)
	}
	if format != detectedUnknown {
		t.Errorf("format = %d, want detectedUnknown", format)
	}
}

func TestDetect_AngleAfterSpace(t *testing.T) {
	// Close angle bracket after space
	data := []byte("<34 >1 test")
	format, err := detect(data)
	if err == nil {
		t.Errorf("expected error for angle bracket after space, got format=%d", format)
	}
}

// --- parserWrapper tests ---

func TestParserWrapper_Dump(t *testing.T) {
	f := &RFC5424{}
	line := []byte("<34>1 2003-10-11T22:14:15Z mymachine su - ID47 - test")
	parser := f.GetParser(line)
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := parser.Dump()
	if parts == nil {
		t.Fatal("Dump returned nil")
	}
	// Verify it returns format.LogParts (map[string]interface{})
	if _, ok := parts["priority"]; !ok {
		t.Error("Dump missing 'priority' key")
	}
}

// --- Integration: Use RFC6587 splitter with bufio.Scanner ---

func TestRFC6587_ScannerIntegration(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	payload := "<34>1 2003-10-11T22:14:15Z host app - - - test"
	data := []byte(strconv.Itoa(len(payload)) + " " + payload + "\n")

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(sf)

	if !scanner.Scan() {
		t.Fatalf("scanner.Scan() returned false: %v", scanner.Err())
	}

	token := scanner.Text()
	if token != payload {
		t.Errorf("scanned token = %q, want %q", token, payload)
	}

	if scanner.Scan() {
		t.Errorf("expected no more tokens, got %q", scanner.Text())
	}
}

// --- Integration: Use Automatic splitter with bufio.Scanner for RFC3164 ---

func TestAutomatic_ScannerIntegration_RFC3164(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	data := []byte("<34>Oct 11 22:14:15 myhost su: test message\n<13>Jan  1 00:00:00 host app: msg\n")

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(sf)

	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 2 {
		t.Errorf("scanned %d lines, want 2", count)
	}
}

// --- Integration: Use Automatic splitter with bufio.Scanner for RFC6587 ---

func TestAutomatic_ScannerIntegration_RFC6587(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	payload := "<34>1 2003-10-11T22:14:15Z host app - - - test"
	data := []byte(strconv.Itoa(len(payload)) + " " + payload)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(sf)

	if !scanner.Scan() {
		t.Fatalf("scanner.Scan() returned false: %v", scanner.Err())
	}

	token := scanner.Text()
	if token != payload {
		t.Errorf("scanned token = %q, want %q", token, payload)
	}
}

// --- RFC6587 splitter: length equals exact data size ---

func TestRFC6587_Splitter_ExactLength(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	payload := "test"
	data := []byte("4 test")

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	if advance != len(data) {
		t.Errorf("advance = %d, want %d", advance, len(data))
	}
	if string(token) != payload {
		t.Errorf("token = %q, want %q", string(token), payload)
	}
}

// --- RFC6587 splitter: length is less than remaining data ---

func TestRFC6587_Splitter_LengthLessThanRemaining(t *testing.T) {
	f := &RFC6587{}
	sf := f.GetSplitFunc()

	// Length 4, but there are more bytes after "test"
	data := []byte("4 testExtraData")

	advance, token, err := sf(data, true)
	if err != nil {
		t.Fatalf("splitter error: %v", err)
	}
	// Should advance past the length prefix + space + payload
	if advance != 6 { // "4 test" = 6 bytes
		t.Errorf("advance = %d, want 6", advance)
	}
	if string(token) != "test" {
		t.Errorf("token = %q, want 'test'", string(token))
	}
}

// --- Automatic splitter: malformed detection error ---

func TestAutomatic_Splitter_DetectionError(t *testing.T) {
	f := &Automatic{}
	sf := f.GetSplitFunc()

	// Data with space but angle bracket after space — causes detection error
	data := []byte("noangle test")
	_, _, err := sf(data, true)
	if err == nil {
		t.Error("expected error for malformed data in automatic splitter")
	}
}
