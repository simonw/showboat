package exec

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestFreePort(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}
	if port == 0 {
		t.Fatal("expected non-zero port")
	}
	// Port should be usable
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("port %d should be available: %v", port, err)
	}
	ln.Close()
}

func TestWaitForPort(t *testing.T) {
	// Start a listener on a free port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	// WaitForPort should succeed quickly
	err = WaitForPort(port, 2*time.Second)
	if err != nil {
		t.Fatalf("expected port %d to be ready: %v", port, err)
	}
}

func TestWaitForPortTimeout(t *testing.T) {
	// Pick a port that nothing is listening on
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}

	// WaitForPort should timeout
	err = WaitForPort(port, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunServer(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}

	// Start a simple Python HTTP server
	proc, err := RunServer("bash", fmt.Sprintf("python3 -m http.server %d", port), "", port, 5*time.Second)
	if err != nil {
		t.Fatalf("RunServer failed: %v", err)
	}
	defer proc.Stop()

	// The server should be listening
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("expected server to be listening on port %d: %v", port, err)
	}
	conn.Close()
}

func TestRunServerWithPORT(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}

	// Use $PORT in the command — RunServer should set PORT env var
	proc, err := RunServer("bash", "python3 -m http.server $PORT", "", port, 5*time.Second)
	if err != nil {
		t.Fatalf("RunServer failed: %v", err)
	}
	defer proc.Stop()

	// The server should be listening on the assigned port
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("expected server to be listening on port %d: %v", port, err)
	}
	conn.Close()
}

func TestRunServerStop(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}

	proc, err := RunServer("bash", fmt.Sprintf("python3 -m http.server %d", port), "", port, 5*time.Second)
	if err != nil {
		t.Fatalf("RunServer failed: %v", err)
	}

	// Stop the server
	proc.Stop()

	// Give the OS a moment to release the port
	time.Sleep(100 * time.Millisecond)

	// Port should no longer be listening
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 200*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatal("expected server to be stopped")
	}
}
