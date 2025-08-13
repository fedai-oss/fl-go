package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TLSConfig represents the TLS configuration for mTLS
type TLSConfig struct {
	Enabled          bool   `yaml:"enabled"`
	CertPath         string `yaml:"cert_path"`
	KeyPath          string `yaml:"key_path"`
	CAPath           string `yaml:"ca_path"`
	ServerName       string `yaml:"server_name"`
	InsecureSkipTLS  bool   `yaml:"insecure_skip_tls"` // For development only
	AutoGenerateCert bool   `yaml:"auto_generate_cert"`
}

// TLSManager handles TLS certificate generation and management
type TLSManager struct {
	config     TLSConfig
	certDir    string
	serverCert tls.Certificate
	clientCert tls.Certificate
	caCert     *x509.Certificate
}

// NewTLSManager creates a new TLS manager
func NewTLSManager(config TLSConfig, certDir string) (*TLSManager, error) {
	tm := &TLSManager{
		config:  config,
		certDir: certDir,
	}

	if config.Enabled {
		if config.AutoGenerateCert {
			if err := tm.generateCertificates(); err != nil {
				return nil, fmt.Errorf("failed to generate certificates: %w", err)
			}
		}
		if err := tm.loadCertificates(); err != nil {
			return nil, fmt.Errorf("failed to load certificates: %w", err)
		}
	}

	return tm, nil
}

// GetServerCredentials returns gRPC server credentials
func (tm *TLSManager) GetServerCredentials() (credentials.TransportCredentials, error) {
	if !tm.config.Enabled {
		return insecure.NewCredentials(), nil
	}

	// Create TLS config for server
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tm.serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    tm.getCertPool(),
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// GetClientCredentials returns gRPC client credentials
func (tm *TLSManager) GetClientCredentials() (credentials.TransportCredentials, error) {
	if !tm.config.Enabled {
		return insecure.NewCredentials(), nil
	}

	// Create TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tm.clientCert},
		RootCAs:      tm.getCertPool(),
		ServerName:   tm.config.ServerName,
		MinVersion:   tls.VersionTLS12,
	}

	if tm.config.InsecureSkipTLS {
		tlsConfig.InsecureSkipVerify = true
	}

	return credentials.NewTLS(tlsConfig), nil
}

// getCertPool returns the CA certificate pool
func (tm *TLSManager) getCertPool() *x509.CertPool {
	if tm.caCert == nil {
		return nil
	}

	pool := x509.NewCertPool()
	pool.AddCert(tm.caCert)
	return pool
}

// generateCertificates generates self-signed certificates for development
func (tm *TLSManager) generateCertificates() error {
	// Ensure certificate directory exists
	if err := os.MkdirAll(tm.certDir, 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Generate CA certificate
	if err := tm.generateCACertificate(); err != nil {
		return fmt.Errorf("failed to generate CA certificate: %w", err)
	}

	// Generate server certificate
	if err := tm.generateServerCertificate(); err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}

	// Generate client certificate
	if err := tm.generateClientCertificate(); err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	return nil
}

// generateCACertificate generates a CA certificate
func (tm *TLSManager) generateCACertificate() error {
	// Generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{"FL-Go"},
			OrganizationalUnit: []string{"Development"},
			Country:            []string{"US"},
			Province:           []string{""},
			Locality:           []string{""},
			CommonName:         "FL-Go CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create(filepath.Join(tm.certDir, "ca.crt"))
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	// Save private key
	keyOut, err := os.Create(filepath.Join(tm.certDir, "ca.key"))
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	return pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes})
}

// generateServerCertificate generates a server certificate signed by the CA
func (tm *TLSManager) generateServerCertificate() error {
	// Load CA certificate and key
	caCert, caKey, err := tm.loadCACertificate()
	if err != nil {
		return err
	}

	// Generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:       []string{"FL-Go"},
			OrganizationalUnit: []string{"Server"},
			Country:            []string{"US"},
			Province:           []string{""},
			Locality:           []string{""},
			CommonName:         "FL-Go Server",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "fl-go-server"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create(filepath.Join(tm.certDir, "server.crt"))
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	// Save private key
	keyOut, err := os.Create(filepath.Join(tm.certDir, "server.key"))
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	return pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes})
}

// generateClientCertificate generates a client certificate signed by the CA
func (tm *TLSManager) generateClientCertificate() error {
	// Load CA certificate and key
	caCert, caKey, err := tm.loadCACertificate()
	if err != nil {
		return err
	}

	// Generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:       []string{"FL-Go"},
			OrganizationalUnit: []string{"Client"},
			Country:            []string{"US"},
			Province:           []string{""},
			Locality:           []string{""},
			CommonName:         "FL-Go Client",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create(filepath.Join(tm.certDir, "client.crt"))
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	// Save private key
	keyOut, err := os.Create(filepath.Join(tm.certDir, "client.key"))
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	return pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes})
}

// loadCACertificate loads the CA certificate and private key
func (tm *TLSManager) loadCACertificate() (*x509.Certificate, interface{}, error) {
	// Load certificate
	certPEM, err := os.ReadFile(filepath.Join(tm.certDir, "ca.crt"))
	if err != nil {
		return nil, nil, err
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// Load private key
	keyPEM, err := os.ReadFile(filepath.Join(tm.certDir, "ca.key"))
	if err != nil {
		return nil, nil, err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// loadCertificates loads the TLS certificates for server and client
func (tm *TLSManager) loadCertificates() error {
	// Load CA certificate
	caCertPath := tm.config.CAPath
	if caCertPath == "" {
		caCertPath = filepath.Join(tm.certDir, "ca.crt")
	}

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return fmt.Errorf("failed to decode CA certificate")
	}

	tm.caCert, err = x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load server certificate
	serverCertPath := tm.config.CertPath
	serverKeyPath := tm.config.KeyPath
	if serverCertPath == "" {
		serverCertPath = filepath.Join(tm.certDir, "server.crt")
	}
	if serverKeyPath == "" {
		serverKeyPath = filepath.Join(tm.certDir, "server.key")
	}

	tm.serverCert, err = tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load client certificate
	clientCertPath := filepath.Join(tm.certDir, "client.crt")
	clientKeyPath := filepath.Join(tm.certDir, "client.key")

	tm.clientCert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	return nil
}

// NewServerOptions returns gRPC server options with mTLS
func (tm *TLSManager) NewServerOptions() ([]grpc.ServerOption, error) {
	creds, err := tm.GetServerCredentials()
	if err != nil {
		return nil, err
	}

	return []grpc.ServerOption{grpc.Creds(creds)}, nil
}

// NewClientDialOptions returns gRPC client dial options with mTLS
func (tm *TLSManager) NewClientDialOptions() ([]grpc.DialOption, error) {
	creds, err := tm.GetClientCredentials()
	if err != nil {
		return nil, err
	}

	return []grpc.DialOption{grpc.WithTransportCredentials(creds)}, nil
}
