//go:build windows

package memprotectopts

import "testing"

func TestReleaseStringValuesKeepsSharedAndUnowned(t *testing.T) {
	secret := string([]byte("release-secret"))
	sharedValue := string([]byte("shared-value"))
	if err := MemProtect(nil, &secret); err != nil {
		t.Fatal(err)
	}
	if err := MemProtect(nil, &sharedValue); err != nil {
		t.Fatal(err)
	}
	m := map[string]any{
		"owned":  secret,
		"shared": sharedValue,
		"plain":  "literal",
		"count":  5,
	}
	if err := ReleaseStringValues(m, map[string]any{"retained": sharedValue}); err != nil {
		t.Fatal(err)
	}
	if m["owned"] != "" {
		t.Fatal("owned protected value was not released")
	}
	if m["shared"] != "shared-value" {
		t.Fatal("value shared with a retained map was released")
	}
	if m["plain"] != "literal" || m["count"] != 5 {
		t.Fatal("unprotected entries were altered")
	}
	if err := MemUnprotectAll(nil); err != nil {
		t.Fatal(err)
	}
}
