package exec

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// ServerProcess represents a running server process that can be stopped.
type ServerProcess struct {
	cmd  *exec.Cmd
	Port int
}

// Stop kills the server process.
func (s *ServerProcess) Stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}

// FreePort returns an available TCP port.
func FreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("finding free port: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

// WaitForPort polls until a TCP connection to the given port succeeds
// or the timeout expires.
func WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("localhost:%d", port)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for port %d", port)
}

// RunServer starts a long-running server process. It sets PORT in the
// environment, starts the command, and waits for the port to become
// available. If waitPort is 0, the server is started without port waiting.
func RunServer(lang, code, workdir string, waitPort int, timeout time.Duration) (*ServerProcess, error) {
	cmd := exec.Command(lang, "-c", code)

	if workdir != "" {
		cmd.Dir = workdir
	}

	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(waitPort))
	cmd.Stdout = os.Stderr // send server output to stderr so it doesn't interfere
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting server: %w", err)
	}

	proc := &ServerProcess{cmd: cmd, Port: waitPort}

	if waitPort > 0 {
		if err := WaitForPort(waitPort, timeout); err != nil {
			proc.Stop()
			return nil, fmt.Errorf("server failed to start: %w", err)
		}
	}

	return proc, nil
}
