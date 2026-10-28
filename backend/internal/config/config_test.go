package config

import (
	"os"
	"reflect"
	"testing"
)

func TestTrustedProxiesEnvironment(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	if err := os.Unsetenv("TRUSTED_PROXIES"); err != nil {
		t.Fatal(err)
	}
	if got, want := Load().TrustedProxies, []string{"127.0.0.1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("default got %v, want %v", got, want)
	}
	t.Setenv("TRUSTED_PROXIES", " 127.0.0.1, ::1,\t10.0.0.0/24\n")
	if got, want := Load().TrustedProxies, []string{"127.0.0.1", "::1", "10.0.0.0/24"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	t.Setenv("TRUSTED_PROXIES", "")
	if len(Load().TrustedProxies) != 0 {
		t.Fatal("empty setting must trust no proxies")
	}
}
