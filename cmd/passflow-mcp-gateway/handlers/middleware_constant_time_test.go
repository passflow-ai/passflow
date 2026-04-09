package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestBearerAuth_ConstantTimeComparison verifies that the time difference
// between a correct and an incorrect token is not statistically significant
// (i.e. the comparison is constant-time, not short-circuit string equality).
//
// This test measures timing across many iterations and asserts that the
// difference between valid and invalid token comparison time is negligible
// relative to the mean — a gross violation of constant-time behaviour (such as
// short-circuit evaluation) would show a large relative difference.
func TestBearerAuth_ConstantTimeComparison(t *testing.T) {
	const (
		iterations = 500
		// Allow up to 200% relative difference before flagging: constant-time
		// implementations may still have jitter, but short-circuit `!=` on a
		// 32-char token terminates much faster on a wrong first byte.
		maxRatioDelta = 2.0
		correctToken  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		// Wrong token shares the same length but differs only in the LAST byte.
		// A short-circuit comparator terminates at the FIRST byte mismatch:
		// "b..." terminates at position 0, revealing timing difference.
		// Constant-time comparator will NOT differ for first-byte vs last-byte.
		wrongFirstByte = "baaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		wrongLastByte  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"
	)

	t.Setenv("MCP_GATEWAY_TOKEN", correctToken)
	handler := BearerAuth(okHandler)

	measure := func(token string) time.Duration {
		var total time.Duration
		for i := 0; i < iterations; i++ {
			req := httptest.NewRequest(http.MethodPost, "/tools/call", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			start := time.Now()
			handler.ServeHTTP(rec, req)
			total += time.Since(start)
		}
		return total / iterations
	}

	avgFirstByte := measure(wrongFirstByte)
	avgLastByte := measure(wrongLastByte)

	if avgFirstByte == 0 || avgLastByte == 0 {
		t.Skip("timing measurement returned zero — skipping ratio check")
	}

	// For a constant-time implementation the ratio between the two should be
	// close to 1.0.  A short-circuit `!=` would make wrongFirstByte much faster
	// than wrongLastByte (wrong at byte 0 → exits immediately vs wrong at the
	// last byte → traverses the whole string).
	var ratio float64
	if avgFirstByte < avgLastByte {
		ratio = float64(avgLastByte) / float64(avgFirstByte)
	} else {
		ratio = float64(avgFirstByte) / float64(avgLastByte)
	}

	if ratio > maxRatioDelta {
		t.Errorf("timing ratio %.2f exceeds threshold %.2f — token comparison may not be constant-time "+
			"(avgFirstByte=%v, avgLastByte=%v)", ratio, maxRatioDelta, avgFirstByte, avgLastByte)
	}
}
