//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	baseURL    = "http://localhost:6001/api"
	tunnelName = "test-tunnel"
)

type TunnelResponse struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	AuthMethod string `json:"auth_method"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port"`
	State      string `json:"state"`
	Error      string `json:"error,omitempty"`
}

type TunnelsResponse struct {
	Tunnels []TunnelResponse `json:"tunnels"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func TestE2E(t *testing.T) {
	// Setup
	cleanup := setupEnvironment(t)
	defer cleanup()

	// Run tests in order
	t.Run("ListTunnelsEmpty", testListTunnelsEmpty)
	t.Run("CreateTunnel", testCreateTunnel)
	t.Run("GetTunnel", testGetTunnel)
	t.Run("StartTunnel", testStartTunnel)
	t.Run("CheckTunnelConnected", testCheckTunnelConnected)
	t.Run("TestTunnelForwarding", testTunnelForwarding)
	t.Run("StopTunnel", testStopTunnel)
	t.Run("CheckTunnelStopped", testCheckTunnelStopped)
	t.Run("DeleteTunnel", testDeleteTunnel)
	t.Run("VerifyDeletion", testVerifyDeletion)
}

func setupEnvironment(t *testing.T) func() {
	t.Helper()

	dir := getTestDir(t)

	// Generate SSH keys
	sshKeyDir := filepath.Join(dir, "ssh-keys")
	configDir := filepath.Join(dir, "config")
	os.MkdirAll(sshKeyDir, 0755)
	os.MkdirAll(configDir, 0755)

	keyPath := filepath.Join(sshKeyDir, "id_rsa")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", keyPath, "-N", "", "-q")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to generate ssh keys: %v", err)
		}
		os.Chmod(keyPath, 0600)
	}

	// Create empty config
	configPath := filepath.Join(configDir, "config.json")
	os.WriteFile(configPath, []byte(`{"tunnels":[]}`), 0644)

	// Start docker compose
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start docker compose: %v", err)
	}

	// Wait for services
	waitForService(t, "ssht", func() bool {
		resp, err := http.Get(baseURL + "/tunnels")
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == 200
	})

	waitForService(t, "ssh-server", func() bool {
		cmd := exec.Command("docker", "compose", "exec", "-T", "ssh-server", "cat", "/config/.ssh/authorized_keys")
		cmd.Dir = dir
		return cmd.Run() == nil
	})

	return func() {
		cmd := exec.Command("docker", "compose", "down", "-v")
		cmd.Dir = dir
		cmd.Run()
	}
}

func getTestDir(t *testing.T) string {
	t.Helper()
	// Get the directory of this test file
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func waitForService(t *testing.T, name string, check func() bool) {
	t.Helper()
	t.Logf("Waiting for %s...", name)
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if check() {
			t.Logf("%s ready", name)
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("%s did not become ready in time", name)
}

func testListTunnelsEmpty(t *testing.T) {
	resp, err := http.Get(baseURL + "/tunnels")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result TunnelsResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Tunnels) != 0 {
		t.Errorf("expected 0 tunnels, got %d", len(result.Tunnels))
	}
}

func testCreateTunnel(t *testing.T) {
	payload := map[string]interface{}{
		"name":        tunnelName,
		"host":        "ssh-server",
		"port":        2222,
		"user":        "testuser",
		"auth_method": "key",
		"key_path":    "/root/.ssh/id_rsa",
		"remote_host": "0.0.0.0",
		"remote_port": 9999,
		"local_host":  "echo-server",
		"local_port":  80,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/tunnels", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var result MessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Message != "tunnel created" {
		t.Errorf("expected 'tunnel created', got %q", result.Message)
	}
}

func testGetTunnel(t *testing.T) {
	resp, err := http.Get(baseURL + "/tunnels/" + tunnelName)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result TunnelResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Name != tunnelName {
		t.Errorf("expected name %q, got %q", tunnelName, result.Name)
	}
	if result.State != "Stopped" {
		t.Errorf("expected state 'Stopped', got %q", result.State)
	}
}

func testStartTunnel(t *testing.T) {
	resp, err := http.Post(baseURL+"/tunnels/"+tunnelName+"/start", "", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result MessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Message != "tunnel started" {
		t.Errorf("expected 'tunnel started', got %q", result.Message)
	}
}

func testCheckTunnelConnected(t *testing.T) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/tunnels/" + tunnelName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		var result TunnelResponse
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if result.State == "Connected" {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Error("tunnel did not connect in time")
}

func testTunnelForwarding(t *testing.T) {
	dir := getTestDir(t)
	cmd := exec.Command("docker", "compose", "exec", "-T", "ssh-server", "curl", "-s", "http://localhost:9999")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("curl failed: %v", err)
	}

	if !bytes.Contains(output, []byte("http")) && !bytes.Contains(output, []byte("echo")) {
		t.Errorf("unexpected response: %s", output)
	}
}

func testStopTunnel(t *testing.T) {
	resp, err := http.Post(baseURL+"/tunnels/"+tunnelName+"/stop", "", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result MessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Message != "tunnel stopped" {
		t.Errorf("expected 'tunnel stopped', got %q", result.Message)
	}
}

func testCheckTunnelStopped(t *testing.T) {
	time.Sleep(1 * time.Second)

	resp, err := http.Get(baseURL + "/tunnels/" + tunnelName)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result TunnelResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.State != "Stopped" {
		t.Errorf("expected state 'Stopped', got %q", result.State)
	}
}

func testDeleteTunnel(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, baseURL+"/tunnels/"+tunnelName, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result MessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Message != "tunnel deleted" {
		t.Errorf("expected 'tunnel deleted', got %q", result.Message)
	}
}

func testVerifyDeletion(t *testing.T) {
	resp, err := http.Get(baseURL + "/tunnels")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result TunnelsResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Tunnels) != 0 {
		t.Errorf("expected 0 tunnels, got %d", len(result.Tunnels))
	}
}
