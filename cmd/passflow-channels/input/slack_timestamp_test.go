package input

import (
	"testing"
	"time"
)

// TestParseTimestamp_ValidUnixSeconds verifies a normal Unix timestamp parses correctly.
func TestParseTimestamp_ValidUnixSeconds(t *testing.T) {
	// 2023-01-01 00:00:00 UTC
	ts, err := parseTimestamp("1672531200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Unix(1672531200, 0)
	if !ts.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, ts)
	}
}

// TestParseTimestamp_NonDigitCharacter verifies that non-digit characters cause an error.
func TestParseTimestamp_NonDigitCharacter(t *testing.T) {
	if _, err := parseTimestamp("169254abc"); err == nil {
		t.Error("expected error for non-digit characters, got nil")
	}
}

// TestParseTimestamp_Empty verifies that an empty string returns an error.
func TestParseTimestamp_Empty(t *testing.T) {
	if _, err := parseTimestamp(""); err == nil {
		t.Error("expected error for empty timestamp, got nil")
	}
}

// TestParseTimestamp_TooFarInFuture verifies that timestamps beyond year 2100
// are rejected to prevent potential overflow / unreasonable values.
func TestParseTimestamp_TooFarInFuture(t *testing.T) {
	// Unix timestamp for year 2101-01-01 (well beyond 2100)
	if _, err := parseTimestamp("4133980800"); err == nil {
		t.Error("expected error for timestamp beyond year 2100, got nil")
	}
}

// TestParseTimestamp_TooFarInPast verifies that timestamps before year 2000
// are rejected.
func TestParseTimestamp_TooFarInPast(t *testing.T) {
	// Unix timestamp for 1999-12-31 (before year 2000)
	if _, err := parseTimestamp("946684799"); err == nil {
		t.Error("expected error for timestamp before year 2000, got nil")
	}
}

// TestParseTimestamp_NegativeValue verifies that a negative-looking string
// (leading non-digit) is rejected.
func TestParseTimestamp_NegativeValue(t *testing.T) {
	if _, err := parseTimestamp("-1672531200"); err == nil {
		t.Error("expected error for negative timestamp string, got nil")
	}
}

// TestParseTimestamp_OverflowLargeString verifies that a very long digit string
// is rejected before it can overflow int64 arithmetic.
func TestParseTimestamp_OverflowLargeString(t *testing.T) {
	// 20 digits — far exceeds int64 max (19 digits)
	if _, err := parseTimestamp("99999999999999999999"); err == nil {
		t.Error("expected error for overflow-length timestamp, got nil")
	}
}

// TestParseTimestamp_BoundaryYear2000 verifies that 2000-01-01 is accepted.
func TestParseTimestamp_BoundaryYear2000(t *testing.T) {
	// 2000-01-01 00:00:00 UTC = 946684800
	if _, err := parseTimestamp("946684800"); err != nil {
		t.Errorf("expected year-2000 boundary to be valid, got: %v", err)
	}
}

// TestParseTimestamp_BoundaryYear2100 verifies that 2100-01-01 is accepted.
func TestParseTimestamp_BoundaryYear2100(t *testing.T) {
	// 2100-01-01 00:00:00 UTC = 4102444800
	if _, err := parseTimestamp("4102444800"); err != nil {
		t.Errorf("expected year-2100 boundary to be valid, got: %v", err)
	}
}
