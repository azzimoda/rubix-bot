package bot

import (
	"net/http"
	"testing"
)

func TestSocks5ClientBuilds(t *testing.T) {
	client, err := socks5Client("127.0.0.1:9050", "", "")
	if err != nil {
		t.Fatalf("socks5Client: %v", err)
	}
	if client == nil {
		t.Fatal("expected a non-nil client")
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if tr.DialContext == nil {
		t.Fatal("expected DialContext to be set on the proxy transport")
	}
}

func TestSocks5ClientWithAuthBuilds(t *testing.T) {
	client, err := socks5Client("127.0.0.1:9050", "user", "pass")
	if err != nil {
		t.Fatalf("socks5Client with auth: %v", err)
	}
	if client == nil {
		t.Fatal("expected a non-nil client")
	}
}
