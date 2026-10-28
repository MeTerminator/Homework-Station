package main

import (
	"bytes"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestOccupiedPortDoesNotLogStartup(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	t.Setenv("PORT", strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)

	err = run()
	var networkError *net.OpError
	if !errors.As(err, &networkError) || networkError.Op != "listen" {
		t.Fatalf("expected listen error, got %v", err)
	}
	if strings.Contains(output.String(), "作业驿站启动") {
		t.Fatal("logged startup even though the port was occupied")
	}
}
