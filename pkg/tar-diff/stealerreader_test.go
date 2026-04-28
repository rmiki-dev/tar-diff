package tardiff

import (
	"bytes"
	"io"
	"testing"
)

func TestNewStealerReader(t *testing.T) {
	source := bytes.NewBufferString("test data")
	var stealer bytes.Buffer

	sr := newStealerReader(source, &stealer)

	if sr == nil {
		t.Fatal("newStealerReader returned nil")
	}

	if sr.source != source {
		t.Error("stealerReader source not set correctly")
	}

	if sr.stealer != &stealer {
		t.Error("stealerReader stealer not set correctly")
	}
}

func TestStealerReaderRead(t *testing.T) {
	testData := []byte("Hello, World!")
	source := bytes.NewBuffer(testData)
	var stealer bytes.Buffer

	sr := newStealerReader(source, &stealer)

	buf := make([]byte, 5)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}

	if string(buf) != "Hello" {
		t.Errorf("Expected to read 'Hello', got %q", string(buf))
	}

	// Check that data was copied to stealer
	stolen := stealer.Bytes()
	if string(stolen) != "Hello" {
		t.Errorf("Expected stealer to contain 'Hello', got %q", string(stolen))
	}
}

func TestStealerReaderSetIgnore(t *testing.T) {
	source := bytes.NewBufferString("test data")
	var stealer bytes.Buffer

	sr := newStealerReader(source, &stealer)

	// Test setting ignore to true
	sr.SetIgnore(true)
	if !sr.ignore {
		t.Error("SetIgnore(true) failed")
	}

	// Test setting ignore to false
	sr.SetIgnore(false)
	if sr.ignore {
		t.Error("SetIgnore(false) failed")
	}
}

func TestStealerReaderReadWithIgnore(t *testing.T) {
	testData := []byte("Hello, World!")
	source := bytes.NewBuffer(testData)
	var stealer bytes.Buffer

	sr := newStealerReader(source, &stealer)
	sr.SetIgnore(true)

	// Read data - should NOT copy to stealer when ignore is true
	buf := make([]byte, 5)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}

	// Check that no data was copied to stealer
	stolen := stealer.Bytes()
	if len(stolen) != 0 {
		t.Errorf("Expected stealer to be empty when ignore=true, got %q", string(stolen))
	}
}

func TestStealerReaderEOF(t *testing.T) {
	testData := []byte("short")
	source := bytes.NewBuffer(testData)
	var stealer bytes.Buffer

	sr := newStealerReader(source, &stealer)

	// Read all data
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}

	// Another read should return EOF
	n2, err := sr.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got error: %v", err)
	}
	if n2 != 0 {
		t.Errorf("Expected 0 bytes on EOF, got %d", n2)
	}
}
