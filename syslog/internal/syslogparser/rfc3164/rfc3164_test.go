package rfc3164

import (
	"testing"
	"time"

	"github.com/tea4go/gh/syslog/internal/syslogparser"
)

// --- NewParser ---

func TestNewParser(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app: msg")
	p := NewParser(buff)
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if string(p.buff) != string(buff) {
		t.Error("buff not set correctly")
	}
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0", p.cursor)
	}
	if p.l != len(buff) {
		t.Errorf("l = %d, want %d", p.l, len(buff))
	}
	if p.location != time.UTC {
		t.Errorf("location = %v, want UTC", p.location)
	}
}

// --- Location ---

func TestParser_Location(t *testing.T) {
	p := NewParser([]byte("test"))
	loc := time.FixedZone("TEST", 3600)
	p.Location(loc)
	if p.location != loc {
		t.Errorf("location not set correctly")
	}
}

// --- Parse: Full valid messages ---

func TestParser_Parse_FullValid(t *testing.T) {
	// Standard RFC3164 message
	buff := []byte("<34>Oct 11 22:14:15 myhost su: 'su root' failed for user on /dev/pts/8")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()
	if parts["priority"] != 34 {
		t.Errorf("priority = %v, want 34", parts["priority"])
	}
	if parts["facility"] != 4 {
		t.Errorf("facility = %v, want 4", parts["facility"])
	}
	if parts["severity"] != 2 {
		t.Errorf("severity = %v, want 2", parts["severity"])
	}
	if parts["hostname"] != "myhost" {
		t.Errorf("hostname = %v, want myhost", parts["hostname"])
	}
	if parts["tag"] != "su" {
		t.Errorf("tag = %v, want su", parts["tag"])
	}
}

func TestParser_Parse_WithPID(t *testing.T) {
	// Message with PID in brackets
	buff := []byte("<13>Jan  1 00:00:00 host app[1234]: message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()
	if parts["tag"] != "app" {
		t.Errorf("tag = %v, want app", parts["tag"])
	}
	if parts["pid"] != 1234 {
		t.Errorf("pid = %v, want 1234", parts["pid"])
	}
	if parts["content"] != "message" {
		t.Errorf("content = %v, want message", parts["content"])
	}
}

func TestParser_Parse_WithoutPID(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app: message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()
	if parts["tag"] != "app" {
		t.Errorf("tag = %v, want app", parts["tag"])
	}
	if parts["pid"] != 0 {
		t.Errorf("pid = %v, want 0", parts["pid"])
	}
}

// --- Parse: Truncated messages ---

func TestParser_Parse_TruncatedNoHeader(t *testing.T) {
	// Message without timestamp - should use current time
	buff := []byte("<13>app: message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()
	// When timestamp is unknown, skipTag is set to true
	if parts["tag"] != "" {
		t.Errorf("tag should be empty when timestamp unknown, got %v", parts["tag"])
	}
	// Check that timestamp is set (approximately now)
	ts, ok := parts["timestamp"].(time.Time)
	if !ok {
		t.Error("timestamp not a time.Time")
	}
	// Should be within a few seconds of now
	now := time.Now()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("timestamp too far from now: %v", diff)
	}
}

func TestParser_Parse_TruncatedPriorityOnly(t *testing.T) {
	buff := []byte("<13>")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		// This is expected to fail or succeed with empty content
		// depending on implementation
		t.Logf("Parse returned: %v", err)
	}
}

// --- Parse: Empty input ---

func TestParser_Parse_Empty(t *testing.T) {
	buff := []byte("")
	p := NewParser(buff)
	err := p.Parse()
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParser_Parse_NoPriorityStart(t *testing.T) {
	buff := []byte("no priority here")
	p := NewParser(buff)
	err := p.Parse()
	if err != syslogparser.ErrPriorityNoStart {
		t.Errorf("expected ErrPriorityNoStart, got %v", err)
	}
}

// --- Parse: PID terminator at buffer end (crash case) ---

func TestParser_Parse_PIDTerminatorAtBufferEnd(t *testing.T) {
	// This tests the fix for the crash when PID terminator is at buffer end
	buff := []byte("<13>Jan  1 00:00:00 host app[1234]:")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()
	if parts["tag"] != "app" {
		t.Errorf("tag = %v, want app", parts["tag"])
	}
	if parts["pid"] != 1234 {
		t.Errorf("pid = %v, want 1234", parts["pid"])
	}
}

func TestParser_Parse_PIDBracketAtBufferEnd(t *testing.T) {
	// PID opening bracket but no closing
	buff := []byte("<13>Jan  1 00:00:00 host app[1234")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Logf("Parse returned: %v", err)
	}
	// Should not panic
}

func TestParser_Parse_PIDEndBracketAtBufferEnd(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app[1234]")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["pid"] != 1234 {
		t.Errorf("pid = %v, want 1234", parts["pid"])
	}
}

// --- Parse: Various timestamp formats ---

func TestParser_Parse_TimestampFormat1(t *testing.T) {
	// "Jan 02 15:04:05" format
	buff := []byte("<13>Jan 01 00:00:00 host app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	ts, ok := parts["timestamp"].(time.Time)
	if !ok {
		t.Fatal("timestamp not a time.Time")
	}
	if ts.Month() != time.January {
		t.Errorf("month = %v, want January", ts.Month())
	}
	if ts.Day() != 1 {
		t.Errorf("day = %v, want 1", ts.Day())
	}
}

func TestParser_Parse_TimestampFormat2(t *testing.T) {
	// "Jan  2 15:04:05" format (single digit day with extra space)
	buff := []byte("<13>Jan  1 00:00:00 host app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	ts, ok := parts["timestamp"].(time.Time)
	if !ok {
		t.Fatal("timestamp not a time.Time")
	}
	if ts.Month() != time.January {
		t.Errorf("month = %v, want January", ts.Month())
	}
	if ts.Day() != 1 {
		t.Errorf("day = %v, want 1", ts.Day())
	}
}

// --- parseTag edge cases ---

func TestParser_Parse_TagWithSpace(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host tag value: content")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	// Tag should be "tag" (stops at space)
	if parts["tag"] != "tag" {
		t.Errorf("tag = %v, want 'tag'", parts["tag"])
	}
}

func TestParser_Parse_NoTag(t *testing.T) {
	// When there's no recognizable tag
	buff := []byte("<13>Jan  1 00:00:00 host ")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Logf("Parse returned: %v", err)
	}
	// Should not panic
}

func TestParser_Parse_InvalidPID(t *testing.T) {
	// PID is not a number — the parser may return an error or default pid to 0;
	// either way it must not panic.
	buff := []byte("<13>Jan  1 00:00:00 host app[abc]: msg")
	p := NewParser(buff)
	_ = p.Parse()
	// If Parse succeeds, pid should be 0 (invalid Atoi defaults to 0)
	parts := p.Dump()
	if pid, ok := parts["pid"]; ok && pid != 0 {
		t.Errorf("pid = %v, want 0 (invalid pid should default to 0)", pid)
	}
}

// --- parseContent ---

func TestParser_Parse_ContentWithSpaces(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app: this is a message with spaces")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["content"] != "this is a message with spaces" {
		t.Errorf("content = %v, want 'this is a message with spaces'", parts["content"])
	}
}

func TestParser_Parse_ContentTrimmed(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app:   message   ")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["content"] != "message" {
		t.Errorf("content = %q, want 'message'", parts["content"])
	}
}

// --- parseHostname ---

func TestParser_Parse_HostnameWithSpace(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 myhost app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["hostname"] != "myhost" {
		t.Errorf("hostname = %v, want 'myhost'", parts["hostname"])
	}
}

// --- Dump ---

func TestParser_Dump_AllFields(t *testing.T) {
	buff := []byte("<165>Jan  1 12:34:56 myhost myapp[5678]: my message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	parts := p.Dump()

	// Check all expected fields exist
	expectedFields := []string{"timestamp", "hostname", "tag", "pid", "content", "priority", "facility", "severity"}
	for _, field := range expectedFields {
		if _, ok := parts[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}

	if parts["priority"] != 165 {
		t.Errorf("priority = %v, want 165", parts["priority"])
	}
	if parts["facility"] != 20 {
		t.Errorf("facility = %v, want 20", parts["facility"])
	}
	if parts["severity"] != 5 {
		t.Errorf("severity = %v, want 5", parts["severity"])
	}
}

// --- fixTimestampIfNeeded ---

func TestFixTimestampIfNeeded(t *testing.T) {
	// Year 0 should be replaced with current year
	ts := time.Date(0, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixTimestampIfNeeded(&ts)
	if ts.Year() == 0 {
		t.Error("year should not be 0 after fixTimestampIfNeeded")
	}
	if ts.Year() != time.Now().Year() {
		t.Errorf("year = %d, want %d", ts.Year(), time.Now().Year())
	}
}

func TestFixTimestampIfNeeded_NonZeroYear(t *testing.T) {
	ts := time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC)
	originalYear := ts.Year()
	fixTimestampIfNeeded(&ts)
	if ts.Year() != originalYear {
		t.Errorf("year changed from %d to %d", originalYear, ts.Year())
	}
}

// --- Edge cases for parseTag ---

func TestParser_Parse_TagEndsWithColon(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host mytag: message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["tag"] != "mytag" {
		t.Errorf("tag = %v, want 'mytag'", parts["tag"])
	}
}

func TestParser_Parse_PIDEndsWithSpace(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app[123] message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["tag"] != "app" {
		t.Errorf("tag = %v, want 'app'", parts["tag"])
	}
	if parts["pid"] != 123 {
		t.Errorf("pid = %v, want 123", parts["pid"])
	}
}

// --- Location affects timestamp parsing ---

func TestParser_Parse_WithLocation(t *testing.T) {
	buff := []byte("<13>Jan  1 12:00:00 host app: msg")
	p := NewParser(buff)
	loc := time.FixedZone("TEST", -5*3600) // UTC-5
	p.Location(loc)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	ts, ok := parts["timestamp"].(time.Time)
	if !ok {
		t.Fatal("timestamp not a time.Time")
	}
	// The timestamp should be in the specified location
	if ts.Location() != loc {
		t.Errorf("timestamp location = %v, want %v", ts.Location(), loc)
	}
}

// --- Priority edge cases ---

func TestParser_Parse_PriorityMax(t *testing.T) {
	buff := []byte("<191>Jan  1 00:00:00 host app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["priority"] != 191 {
		t.Errorf("priority = %v, want 191", parts["priority"])
	}
}

func TestParser_Parse_PriorityMin(t *testing.T) {
	buff := []byte("<0>Jan  1 00:00:00 host app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["priority"] != 0 {
		t.Errorf("priority = %v, want 0", parts["priority"])
	}
}

// --- Message with special characters ---

func TestParser_Parse_SpecialChars(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app: message with <special> chars & symbols!")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	expected := "message with <special> chars & symbols!"
	if parts["content"] != expected {
		t.Errorf("content = %v, want %v", parts["content"], expected)
	}
}

// --- Cursor position tests ---

func TestParser_CursorAfterParse(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app: msg")
	p := NewParser(buff)
	_ = p.Parse()
	// Cursor should be at end after successful parse
	if p.cursor != p.l {
		t.Errorf("cursor = %d, want %d", p.cursor, p.l)
	}
}

// --- Empty content ---

func TestParser_Parse_EmptyContent(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app:")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["content"] != "" {
		t.Errorf("content = %v, want empty", parts["content"])
	}
}

// --- Multiple spaces in timestamp ---

func TestParser_Parse_MultipleSpacesInTimestamp(t *testing.T) {
	// Single digit day with double space
	buff := []byte("<13>Jan  5 10:30:00 host app: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	ts, ok := parts["timestamp"].(time.Time)
	if !ok {
		t.Fatal("timestamp not a time.Time")
	}
	if ts.Day() != 5 {
		t.Errorf("day = %d, want 5", ts.Day())
	}
}

// --- Tag with numbers ---

func TestParser_Parse_TagWithNumbers(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app2[123]: msg")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	if parts["tag"] != "app2" {
		t.Errorf("tag = %v, want 'app2'", parts["tag"])
	}
}

// --- No colon after tag ---

func TestParser_Parse_NoColonAfterTag(t *testing.T) {
	buff := []byte("<13>Jan  1 00:00:00 host app message")
	p := NewParser(buff)
	err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	parts := p.Dump()
	// Tag stops at space
	if parts["tag"] != "app" {
		t.Errorf("tag = %v, want 'app'", parts["tag"])
	}
}
