package service

import "testing"

func TestIdentityJSONFields(t *testing.T) {
	if !jsonMatches(true, "true") || jsonMatches("true", "true") || !jsonMatches("ok", `"ok"`) || jsonMatches(nil, "true") {
		t.Fatal("success comparison must preserve JSON types")
	}
	if got := realNameURL("https://example.com/{name}?id={id}&name={name}", "2026000001", "A B&+#"); got != "https://example.com/A%20B&+%23?id=2026000001&name=A+B%26%2B%23" {
		t.Fatal(got)
	}
}
