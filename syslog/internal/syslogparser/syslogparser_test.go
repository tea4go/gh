package syslogparser

import (
	"testing"
	"time"
)

// --- ParserError ---

func TestParserError_Error(t *testing.T) {
	err := &ParserError{ErrorString: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
}

// --- IsDigit ---

func TestIsDigit(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'0', true},
		{'5', true},
		{'9', true},
		{'a', false},
		{'Z', false},
		{'/', false},
		{':', false},
		{' ', false},
		{0, false},
		{255, false},
	}
	for _, tt := range tests {
		if got := IsDigit(tt.input); got != tt.want {
			t.Errorf("IsDigit(%d) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- newPriority ---

func TestNewPriority(t *testing.T) {
	p := newPriority(0)
	if p.P != 0 || p.F.Value != 0 || p.S.Value != 0 {
		t.Errorf("newPriority(0) = %+v, unexpected", p)
	}

	p = newPriority(191)
	if p.P != 191 {
		t.Errorf("P = %d, want 191", p.P)
	}
	if p.F.Value != 191/8 {
		t.Errorf("F.Value = %d, want %d", p.F.Value, 191/8)
	}
	if p.S.Value != 191%8 {
		t.Errorf("S.Value = %d, want %d", p.S.Value, 191%8)
	}

	p = newPriority(13)
	if p.F.Value != 1 {
		t.Errorf("F.Value = %d, want 1", p.F.Value)
	}
	if p.S.Value != 5 {
		t.Errorf("S.Value = %d, want 5", p.S.Value)
	}
}

// --- ParsePriority ---

func TestParsePriority_Valid(t *testing.T) {
	buff := []byte("<123>")
	cursor := 0
	pri, err := ParsePriority(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pri.P != 123 {
		t.Errorf("P = %d, want 123", pri.P)
	}
	if cursor != 5 {
		t.Errorf("cursor = %d, want 5", cursor)
	}
}

func TestParsePriority_SingleDigit(t *testing.T) {
	buff := []byte("<5>")
	cursor := 0
	pri, err := ParsePriority(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pri.P != 5 {
		t.Errorf("P = %d, want 5", pri.P)
	}
}

func TestParsePriority_Zero(t *testing.T) {
	buff := []byte("<0>rest")
	cursor := 0
	pri, err := ParsePriority(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pri.P != 0 {
		t.Errorf("P = %d, want 0", pri.P)
	}
	if cursor != 3 {
		t.Errorf("cursor = %d, want 3", cursor)
	}
}

func TestParsePriority_EmptyBuffer(t *testing.T) {
	buff := []byte{}
	cursor := 0
	_, err := ParsePriority(buff, &cursor, 0)
	if err != ErrPriorityEmpty {
		t.Errorf("expected ErrPriorityEmpty, got %v", err)
	}
}

func TestParsePriority_NoStart(t *testing.T) {
	buff := []byte("123>")
	cursor := 0
	_, err := ParsePriority(buff, &cursor, len(buff))
	if err != ErrPriorityNoStart {
		t.Errorf("expected ErrPriorityNoStart, got %v", err)
	}
}

func TestParsePriority_TooShort(t *testing.T) {
	buff := []byte("<>")
	cursor := 0
	_, err := ParsePriority(buff, &cursor, len(buff))
	if err != ErrPriorityTooShort {
		t.Errorf("expected ErrPriorityTooShort, got %v", err)
	}
}

func TestParsePriority_NoEnd(t *testing.T) {
	buff := []byte("<123")
	cursor := 0
	_, err := ParsePriority(buff, &cursor, len(buff))
	if err != ErrPriorityNoEnd {
		t.Errorf("expected ErrPriorityNoEnd, got %v", err)
	}
}

func TestParsePriority_TooLong(t *testing.T) {
	buff := []byte("<12345>")
	cursor := 0
	_, err := ParsePriority(buff, &cursor, len(buff))
	if err != ErrPriorityTooLong {
		t.Errorf("expected ErrPriorityTooLong, got %v", err)
	}
}

func TestParsePriority_NonDigit(t *testing.T) {
	buff := []byte("<1a3>")
	cursor := 0
	_, err := ParsePriority(buff, &cursor, len(buff))
	if err != ErrPriorityNonDigit {
		t.Errorf("expected ErrPriorityNonDigit, got %v", err)
	}
}

// --- ParseVersion ---

func TestParseVersion_Valid(t *testing.T) {
	buff := []byte("1 rest")
	cursor := 0
	v, err := ParseVersion(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1 {
		t.Errorf("version = %d, want 1", v)
	}
	if cursor != 1 {
		t.Errorf("cursor = %d, want 1", cursor)
	}
}

func TestParseVersion_NotDigit(t *testing.T) {
	// When the character is not a digit, it should return NO_VERSION and no error
	buff := []byte("A rest")
	cursor := 0
	v, err := ParseVersion(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != NO_VERSION {
		t.Errorf("version = %d, want NO_VERSION (%d)", v, NO_VERSION)
	}
}

func TestParseVersion_CursorAtEnd(t *testing.T) {
	buff := []byte("1")
	cursor := 1 // at end
	v, err := ParseVersion(buff, &cursor, len(buff))
	if err != ErrVersionNotFound {
		t.Errorf("expected ErrVersionNotFound, got %v", err)
	}
	if v != NO_VERSION {
		t.Errorf("version = %d, want NO_VERSION", v)
	}
}

// --- FindNextSpace ---

func TestFindNextSpace_Found(t *testing.T) {
	buff := []byte("abc def")
	from := 0
	to, err := FindNextSpace(buff, from, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if to != 4 {
		t.Errorf("to = %d, want 4", to)
	}
}

func TestFindNextSpace_NotFound(t *testing.T) {
	buff := []byte("abcdef")
	from := 0
	_, err := FindNextSpace(buff, from, len(buff))
	if err != ErrNoSpace {
		t.Errorf("expected ErrNoSpace, got %v", err)
	}
}

func TestFindNextSpace_FromMiddle(t *testing.T) {
	buff := []byte("abc def ghi")
	from := 4
	to, err := FindNextSpace(buff, from, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if to != 8 {
		t.Errorf("to = %d, want 8", to)
	}
}

// --- Parse2Digits ---

func TestParse2Digits_Valid(t *testing.T) {
	buff := []byte("12")
	cursor := 0
	v, err := Parse2Digits(buff, &cursor, len(buff), 0, 99, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 12 {
		t.Errorf("value = %d, want 12", v)
	}
	if cursor != 2 {
		t.Errorf("cursor = %d, want 2", cursor)
	}
}

func TestParse2Digits_OutOfRange(t *testing.T) {
	buff := []byte("99")
	cursor := 0
	customErr := &ParserError{ErrorString: "out of range"}
	_, err := Parse2Digits(buff, &cursor, len(buff), 0, 50, customErr)
	if err != customErr {
		t.Errorf("expected customErr, got %v", err)
	}
}

func TestParse2Digits_BufferTooShort(t *testing.T) {
	buff := []byte("1")
	cursor := 0
	_, err := Parse2Digits(buff, &cursor, len(buff), 0, 99, nil)
	if err != ErrEOL {
		t.Errorf("expected ErrEOL, got %v", err)
	}
}

func TestParse2Digits_NonDigit(t *testing.T) {
	buff := []byte("ab")
	cursor := 0
	customErr := &ParserError{ErrorString: "custom"}
	_, err := Parse2Digits(buff, &cursor, len(buff), 0, 99, customErr)
	if err != customErr {
		t.Errorf("expected customErr, got %v", err)
	}
}

// --- ParseHostname ---

func TestParseHostname_Valid(t *testing.T) {
	buff := []byte("myhost rest")
	cursor := 0
	hostname, err := ParseHostname(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "myhost" {
		t.Errorf("hostname = %q, want %q", hostname, "myhost")
	}
	if cursor != 6 {
		t.Errorf("cursor = %d, want 6 (stops at space)", cursor)
	}
}

func TestParseHostname_NoSpace(t *testing.T) {
	buff := []byte("myhost")
	cursor := 0
	hostname, err := ParseHostname(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "myhost" {
		t.Errorf("hostname = %q, want %q", hostname, "myhost")
	}
	if cursor != 6 {
		t.Errorf("cursor = %d, want 6", cursor)
	}
}

func TestParseHostname_Empty(t *testing.T) {
	buff := []byte(" rest")
	cursor := 0
	hostname, err := ParseHostname(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "" {
		t.Errorf("hostname = %q, want empty", hostname)
	}
}

// --- ShowCursorPos (just ensure it doesn't panic) ---

func TestShowCursorPos(t *testing.T) {
	ShowCursorPos([]byte("hello"), 2)
}

// --- LogParser interface compile check ---

func TestLogPartsType(t *testing.T) {
	lp := LogParts{"key": "value"}
	if lp["key"] != "value" {
		t.Errorf("LogParts map not working as expected")
	}
}

// --- Priority/Facility/Severity structs ---

func TestFacilitySeverity(t *testing.T) {
	f := Facility{Value: 23}
	s := Severity{Value: 5}
	if f.Value != 23 || s.Value != 5 {
		t.Errorf("Facility/Severity struct values incorrect")
	}
}

// --- Integration-style: ParsePriority with cursor offset ---

func TestParsePriority_CursorOffset(t *testing.T) {
	// ParsePriority starts from buff[0], not *cursor, so a non-zero cursor
	// pointing to '<' is not how the API works — buff must start with '<'.
	// Test with buff starting at '<' directly.
	buff := []byte("<42>rest")
	cursor := 0
	pri, err := ParsePriority(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pri.P != 42 {
		t.Errorf("P = %d, want 42", pri.P)
	}
	if cursor != 4 {
		t.Errorf("cursor = %d, want 4", cursor)
	}
}

// --- Location interface check ---

func TestLogParserInterface(t *testing.T) {
	// Just verify the interface exists and has the right methods
	var _ LogParser = (*mockParser)(nil)
}

type mockParser struct{}

func (m *mockParser) Parse() error                    { return nil }
func (m *mockParser) Dump() LogParts                  { return LogParts{} }
func (m *mockParser) Location(loc *time.Location)     {}

// --- ParseVersion edge: non-digit after cursor moves back ---

func TestParseVersion_NonDigitThenAtoi(t *testing.T) {
	// This path is hard to trigger since IsDigit guards it,
	// but we verify the logic path
	buff := []byte("1rest")
	cursor := 0
	v, err := ParseVersion(buff, &cursor, len(buff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1 {
		t.Errorf("version = %d, want 1", v)
	}
}

// --- Parse2Digits boundary values ---

func TestParse2Digits_ExactMin(t *testing.T) {
	buff := []byte("00")
	cursor := 0
	v, err := Parse2Digits(buff, &cursor, len(buff), 0, 99, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("value = %d, want 0", v)
	}
}

func TestParse2Digits_ExactMax(t *testing.T) {
	buff := []byte("23")
	cursor := 0
	v, err := Parse2Digits(buff, &cursor, len(buff), 0, 23, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 23 {
		t.Errorf("value = %d, want 23", v)
	}
}
