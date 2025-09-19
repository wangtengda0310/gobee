package sign

import (
	"testing"
)

// TestSimpleMd5 is a simple test to verify the MD5 signing functionality
func TestSimpleMd5(t *testing.T) {
	// Create a test struct with sign tags
	testStruct := &struct {
		Name      string `sign:"yes"`
		Value     int    `sign:"yes"`
		Signature string `sign:"access"`
	}{
		Name:  "TestValue",
		Value: 123,
	}

	// Apply the MD5 signature
	Md5Apply("test_key", testStruct)

	// Verify that the signature field was set
	if testStruct.Signature == "" {
		t.Error("Signature field was not set")
	}

	// Get the expected signature using the Md5 function
	expected := Md5("test_key", testStruct)

	// Verify that the signature matches the expected value
	if testStruct.Signature != expected {
		t.Errorf("Expected signature %s, got %s", expected, testStruct.Signature)
	}

	t.Logf("Successfully verified MD5 signature: %s", testStruct.Signature)
}
