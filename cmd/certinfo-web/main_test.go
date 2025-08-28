package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCertInfoGET tests the GET request handler
func TestCertInfoGET(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CertInfo)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains expected HTML elements
	body := rr.Body.String()
	expectedElements := []string{
		"<title>Display X.509 certificate info</title>",
		"<h1>X.509 Certificate Information</h1>",
		"<input type=\"file\"",
		"<button type=\"submit\">",
	}

	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("Expected response to contain '%s'", element)
		}
	}
}

// TestCertInfoPOST tests the POST request handler with invalid data
func TestCertInfoPOST(t *testing.T) {
	// Create a test request with invalid form data
	req, err := http.NewRequest("POST", "/", strings.NewReader("invalid=form"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CertInfo)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check that we get an error status (since we're not providing a proper file)
	if status := rr.Code; status == http.StatusOK {
		t.Errorf("Expected error status for invalid form data, got %v", status)
	}
}

// TestTemplateExists tests that the template is properly loaded
func TestTemplateExists(t *testing.T) {
	if templ == nil {
		t.Error("Template should not be nil")
	}
}
