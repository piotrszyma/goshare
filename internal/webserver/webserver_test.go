package webserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestRenderIndexTemplate tests the renderIndexTemplate function
func TestRenderIndexTemplate(t *testing.T) {
	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test case 1: Basic template rendering with no files
	err := renderIndexTemplate(rr, req, "", "")
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains expected elements
	response := rr.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}

	// Check that the response contains the title
	body := rr.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	// Check that the response contains the title
	if !bytes.Contains(rr.Body.Bytes(), []byte("<title>GoShare - File Sharing</title>")) {
		t.Error("Expected response to contain the page title")
	}

	// Check that the response contains the key
	if !bytes.Contains(rr.Body.Bytes(), []byte(`action="/upload?key=`)) {
		t.Error("Expected response to contain the key in the form action")
	}
}

// TestRenderIndexTemplateWithMessage tests template rendering with a message
func TestRenderIndexTemplateWithMessage(t *testing.T) {
	// Create a test request with message parameters
	req := httptest.NewRequest("GET", "/?message=Test%20message&type=success", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with a message
	err := renderIndexTemplate(rr, req, "", "")
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains the message
	if !bytes.Contains(rr.Body.Bytes(), []byte("Test message")) {
		t.Error("Expected response to contain the test message")
	}

	// Check that the response contains the success class
	if !bytes.Contains(rr.Body.Bytes(), []byte("success")) {
		t.Error("Expected response to contain the success class")
	}
}

// TestRenderIndexTemplateWithError tests template rendering with an error message
func TestRenderIndexTemplateWithError(t *testing.T) {
	// Create a test request with error message parameters
	req := httptest.NewRequest("GET", "/?message=Error%20message&type=error", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with an error message
	err := renderIndexTemplate(rr, req, "", "")
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains the error message
	if !bytes.Contains(rr.Body.Bytes(), []byte("Error message")) {
		t.Error("Expected response to contain the error message")
	}

	// Check that the response contains the error class
	if !bytes.Contains(rr.Body.Bytes(), []byte("error")) {
		t.Error("Expected response to contain the error class")
	}
}

// TestRenderIndexTemplateWithDefaultMessageType tests template rendering with a message but no type
func TestRenderIndexTemplateWithDefaultMessageType(t *testing.T) {
	// Create a test request with message but no type
	req := httptest.NewRequest("GET", "/?message=Test%20message", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with a message but no type
	err := renderIndexTemplate(rr, req, "", "")
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains the message
	if !bytes.Contains(rr.Body.Bytes(), []byte("Test message")) {
		t.Error("Expected response to contain the test message")
	}

	// Check that the response contains the default success class
	if !bytes.Contains(rr.Body.Bytes(), []byte("success")) {
		t.Error("Expected response to contain the default success class")
	}
}

// TestRenderIndexTemplateWithUploadsFiles tests template rendering with uploads files
func TestRenderIndexTemplateWithUploadsFiles(t *testing.T) {
	// Create a temporary directory for uploads
	tmpDir, err := os.MkdirTemp("", "uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file in the uploads directory
	testFile, err := os.Create(tmpDir + "/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	testFile.WriteString("test content")
	testFile.Close()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with uploads files
	err = renderIndexTemplate(rr, req, tmpDir, "")
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains the file name
	if !bytes.Contains(rr.Body.Bytes(), []byte("test.txt")) {
		t.Error("Expected response to contain the test file name")
	}

	// Check that the response contains the file size
	if !bytes.Contains(rr.Body.Bytes(), []byte("12 bytes")) {
		t.Error("Expected response to contain the test file size")
	}

	// Check that the response contains the uploads heading
	if !bytes.Contains(rr.Body.Bytes(), []byte("Uploaded Files")) {
		t.Error("Expected response to contain the uploads heading")
	}
}

// TestRenderIndexTemplateWithSharedFiles tests template rendering with shared files
func TestRenderIndexTemplateWithSharedFiles(t *testing.T) {
	// Create a temporary directory for shared files
	tmpDir, err := os.MkdirTemp("", "shared")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file in the shared directory
	testFile, err := os.Create(tmpDir + "/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	testFile.WriteString("shared content")
	testFile.Close()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with shared files
	err = renderIndexTemplate(rr, req, "", tmpDir)
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains the file name
	if !bytes.Contains(rr.Body.Bytes(), []byte("shared.txt")) {
		t.Error("Expected response to contain the shared file name")
	}

	// Check that the response contains the file size
	if !bytes.Contains(rr.Body.Bytes(), []byte("14 bytes")) {
		t.Error("Expected response to contain the shared file size")
	}

	// Check that the response contains the shared heading
	if !bytes.Contains(rr.Body.Bytes(), []byte("Shared Files")) {
		t.Error("Expected response to contain the shared heading")
	}
}

// TestRenderIndexTemplateWithBothFileTypes tests template rendering with both uploads and shared files
func TestRenderIndexTemplateWithBothFileTypes(t *testing.T) {
	// Create temporary directories for uploads and shared files
	uploadsDir, err := os.MkdirTemp("", "uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(uploadsDir)

	sharedDir, err := os.MkdirTemp("", "shared")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sharedDir)

	// Create test files in both directories
	uploadsFile, err := os.Create(uploadsDir + "/upload.txt")
	if err != nil {
		t.Fatal(err)
	}
	uploadsFile.WriteString("upload content")
	uploadsFile.Close()

	sharedFile, err := os.Create(sharedDir + "/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	sharedFile.WriteString("shared content")
	sharedFile.Close()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Test template rendering with both file types
	err = renderIndexTemplate(rr, req, uploadsDir, sharedDir)
	if err != nil {
		t.Errorf("renderIndexTemplate returned an error: %v", err)
	}

	// Check that the response contains both file names
	if !bytes.Contains(rr.Body.Bytes(), []byte("upload.txt")) {
		t.Error("Expected response to contain the upload file name")
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("shared.txt")) {
		t.Error("Expected response to contain the shared file name")
	}

	// Check that the response contains both headings
	if !bytes.Contains(rr.Body.Bytes(), []byte("Uploaded Files")) {
		t.Error("Expected response to contain the uploads heading")
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("Shared Files")) {
		t.Error("Expected response to contain the shared heading")
	}
}
