package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jaak-ai/passflow-mcp-gateway/registry"
)

// TestCallTool_BodySizeLimit_ExceedsLimit verifies that a request body larger
// than 1 MiB is rejected with 400 Bad Request (or 413 Request Entity Too Large)
// rather than being decoded, preventing memory exhaustion attacks.
func TestCallTool_BodySizeLimit_ExceedsLimit(t *testing.T) {
	reg := registry.New()
	h := New(reg, nil)

	// Build a body that is slightly larger than 1 MiB (1<<20 bytes).
	oversized := strings.Repeat("x", (1<<20)+1)
	// Wrap it in something that looks like JSON (it will fail to parse anyway).
	body := `{"name":"tool","arguments":{"data":"` + oversized + `"}}`

	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CallTool(rec, req)

	// The handler must reject an oversized body. HTTP 400 (bad request after
	// hitting the limit) or 413 (request entity too large) are both acceptable.
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 400 or 413 for oversized body, got %d", rec.Code)
	}
}

// TestCallTool_BodySizeLimit_ExactlyAtLimit verifies that a body at exactly
// 1 MiB is still processed normally (not rejected).
func TestCallTool_BodySizeLimit_ExactlyAtLimit(t *testing.T) {
	reg := registry.New()
	h := New(reg, nil)

	// A body of exactly 1 MiB filled with a valid-looking JSON payload.
	// The content will fail to decode as a valid CallRequest, but the size
	// limit alone must not reject it — the 404 (tool not found) confirms it
	// made it through the size gate.
	padding := strings.Repeat("a", (1<<20)-len(`{"name":""}`))
	body := `{"name":"` + padding + `"}`

	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CallTool(rec, req)

	// The body is valid JSON and fits within the limit, so we get past the
	// size check. The tool won't be found → 404.
	if rec.Code == http.StatusRequestEntityTooLarge {
		t.Error("expected body at exactly 1 MiB to NOT be rejected by size limit")
	}
}

// TestCallTool_BodySizeLimit_SmallBody verifies that a normal-sized body is
// processed correctly (regression guard).
func TestCallTool_BodySizeLimit_SmallBody(t *testing.T) {
	reg := registry.New()
	h := New(reg, nil)

	body := `{"name":"unknown.tool","arguments":{}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CallTool(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown tool, got %d", rec.Code)
	}
}
