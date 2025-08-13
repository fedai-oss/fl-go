package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTLSManager_AutoGenerateCert(t *testing.T) {
	// Create temporary directory for certificates
	tempDir := t.TempDir()

	config := TLSConfig{
		Enabled:          true,
		AutoGenerateCert: true,
		ServerName:       "test-server",
		InsecureSkipTLS:  true, // For testing
	}

	// Test TLS manager creation with auto-generated certificates
	tlsManager, err := NewTLSManager(config, tempDir)
	if err != nil {
		t.Fatalf("Failed to create TLS manager: %v", err)
	}

	// Verify certificate files were created
	expectedFiles := []string{"ca.crt", "ca.key", "server.crt", "server.key", "client.crt", "client.key"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(tempDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected certificate file %s was not created", file)
		}
	}

	// Test getting server credentials
	_, err = tlsManager.GetServerCredentials()
	if err != nil {
		t.Errorf("Failed to get server credentials: %v", err)
	}

	// Test getting client credentials
	_, err = tlsManager.GetClientCredentials()
	if err != nil {
		t.Errorf("Failed to get client credentials: %v", err)
	}
}

func TestTLSManager_DisabledTLS(t *testing.T) {
	config := TLSConfig{
		Enabled: false,
	}

	tlsManager, err := NewTLSManager(config, "")
	if err != nil {
		t.Fatalf("Failed to create TLS manager with disabled TLS: %v", err)
	}

	// Test that insecure credentials are returned when TLS is disabled
	serverCreds, err := tlsManager.GetServerCredentials()
	if err != nil {
		t.Errorf("Failed to get server credentials: %v", err)
	}

	// Should not be nil even when TLS is disabled (returns insecure credentials)
	if serverCreds == nil {
		t.Error("Server credentials should not be nil")
	}

	clientCreds, err := tlsManager.GetClientCredentials()
	if err != nil {
		t.Errorf("Failed to get client credentials: %v", err)
	}

	if clientCreds == nil {
		t.Error("Client credentials should not be nil")
	}
}

func TestTLSManager_ServerClientOptions(t *testing.T) {
	tempDir := t.TempDir()

	config := TLSConfig{
		Enabled:          true,
		AutoGenerateCert: true,
		ServerName:       "test-server",
		InsecureSkipTLS:  true,
	}

	tlsManager, err := NewTLSManager(config, tempDir)
	if err != nil {
		t.Fatalf("Failed to create TLS manager: %v", err)
	}

	// Test server options
	serverOpts, err := tlsManager.NewServerOptions()
	if err != nil {
		t.Errorf("Failed to get server options: %v", err)
	}

	if len(serverOpts) == 0 {
		t.Error("Server options should not be empty when TLS is enabled")
	}

	// Test client dial options
	clientOpts, err := tlsManager.NewClientDialOptions()
	if err != nil {
		t.Errorf("Failed to get client dial options: %v", err)
	}

	if len(clientOpts) == 0 {
		t.Error("Client dial options should not be empty when TLS is enabled")
	}
}

func TestTLSConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  TLSConfig
		wantErr bool
	}{
		{
			name: "valid config with auto-generated certs",
			config: TLSConfig{
				Enabled:          true,
				AutoGenerateCert: true,
				ServerName:       "test-server",
			},
			wantErr: false,
		},
		{
			name: "disabled TLS",
			config: TLSConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled TLS with custom paths",
			config: TLSConfig{
				Enabled:    true,
				CertPath:   "/path/to/cert.pem",
				KeyPath:    "/path/to/key.pem",
				CAPath:     "/path/to/ca.pem",
				ServerName: "custom-server",
			},
			wantErr: true, // Will fail because files don't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			_, err := NewTLSManager(tt.config, tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTLSManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
