package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

// TestReadPEMBlocks tests the ReadPEMBlocks function with a simple certificate
func TestReadPEMBlocks(t *testing.T) {
	// Create a simple test certificate
	cert := createTestCertificate(t)

	// Encode as PEM
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Test ReadPEMBlocks
	blocks := ReadPEMBlocks(pemData)

	if len(blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(blocks))
	}

	if len(blocks[0]) == 0 {
		t.Error("Expected non-empty block data")
	}
}

// TestTryParseCert tests the TryParseCert function
func TestTryParseCert(t *testing.T) {
	// Create a test certificate
	cert := createTestCertificate(t)

	// Test parsing the certificate
	parsedCert, err := TryParseCert(cert.Raw)
	if err != nil {
		t.Errorf("Failed to parse certificate: %v", err)
	}

	if parsedCert == nil {
		t.Error("Expected non-nil certificate")
	}

	// Verify basic certificate properties
	if parsedCert.Subject.CommonName != "test.example.com" {
		t.Errorf("Expected CommonName 'test.example.com', got '%s'", parsedCert.Subject.CommonName)
	}
}

// TestGetCertInfoString tests that the certificate info string contains expected information
func TestGetCertInfoString(t *testing.T) {
	cert := createTestCertificate(t)

	info := GetCertInfoString(cert)

	// Check that the info string contains expected fields
	expectedFields := []string{
		"Subject:",
		"Issuer:",
		"Serial:",
		"Version:",
		"Signature Algorithm:",
		"Public Key:",
		"Validity:",
		"Not Before:",
		"Not After:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(info, field) {
			t.Errorf("Expected info string to contain '%s'", field)
		}
	}

	// Check that it contains the test subject
	if !strings.Contains(info, "test.example.com") {
		t.Errorf("Expected info string to contain 'test.example.com'")
	}
}

// TestHexColon tests the hexColon helper function
func TestHexColon(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected string
	}{
		{[]byte{0x01, 0x02, 0x03}, "01:02:03"},
		{[]byte{0xFF, 0xFE}, "FF:FE"},
		{[]byte{0x00}, "00"},
		{[]byte{}, ""},
	}

	for _, tc := range testCases {
		result := hexColon(tc.input)
		if result != tc.expected {
			t.Errorf("hexColon(%v) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// Helper function to create a test certificate
func createTestCertificate(t *testing.T) *x509.Certificate {
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Parse the certificate back
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}
