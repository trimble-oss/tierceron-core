package memprotectopts

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLockedReadAllGrowsAcrossAllocations(t *testing.T) {
	payload := strings.Repeat("a", lockedReadInitialSize*3+17)
	got, err := LockedReadAll(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer LockedFree(got)
	if string(got) != payload {
		t.Fatal("locked read returned different content")
	}
}

func TestExtractLockedStringsRedactsTokens(t *testing.T) {
	body := []byte(`{"token_type":"Bearer","access_token":"secret-access","nested":{"refresh_token":"secret-refresh"},"id_token":"with\u0020escape","expires_in":3600}`)
	extracted, err := ExtractLockedStrings(body, "access_token", "refresh_token", "id_token")
	if err != nil {
		t.Fatal(err)
	}
	if extracted["access_token"] != "secret-access" || extracted["refresh_token"] != "secret-refresh" {
		t.Fatalf("unexpected extraction: %q", extracted)
	}
	if _, ok := extracted["id_token"]; ok {
		t.Fatal("escaped value must fall through to the decoder")
	}
	if bytes.Contains(body, []byte("secret-access")) || bytes.Contains(body, []byte("secret-refresh")) {
		t.Fatal("token bytes were not redacted from the body")
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("redacted body is no longer valid JSON: %v", err)
	}
	if decoded["token_type"] != "Bearer" || decoded["expires_in"] != float64(3600) {
		t.Fatal("redaction altered non-secret fields")
	}
	if decoded["id_token"] != "with escape" {
		t.Fatal("escaped value must survive for decoder fallback")
	}
	if !strings.Contains(decoded["access_token"].(string), "X") {
		t.Fatal("decoded body should only contain filler for extracted tokens")
	}
}

func TestExtractLockedStringsIgnoresKeyLikeValues(t *testing.T) {
	body := []byte(`{"note":"access_token","access_token":"real-secret"}`)
	extracted, err := ExtractLockedStrings(body, "access_token")
	if err != nil {
		t.Fatal(err)
	}
	if extracted["access_token"] != "real-secret" {
		t.Fatalf("value resembling a key confused extraction: %q", extracted)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["note"] != "access_token" {
		t.Fatal("redaction altered an unrelated value")
	}
}
