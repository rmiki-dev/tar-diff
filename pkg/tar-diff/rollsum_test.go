package tardiff

import (
	"hash/crc32"
	"testing"
)

func TestNewRollsum(t *testing.T) {
	r := newRollsum()

	if r == nil {
		t.Fatal("newRollsum() returned nil")
	}

	if len(r.header) != 0 {
		t.Errorf("Expected header length 0, got %d", len(r.header))
	}

	if cap(r.header) != 16 {
		t.Errorf("Expected header capacity 16, got %d", cap(r.header))
	}
}

func TestRollsumInit(t *testing.T) {
	r := newRollsum()
	r.blobSize = 100

	r.init()

	if r.blobStart != 100 {
		t.Errorf("Expected blobStart 100, got %d", r.blobStart)
	}

	if r.blobSize != 0 {
		t.Errorf("Expected blobSize 0 after init, got %d", r.blobSize)
	}
}

func TestRollsumWrite(t *testing.T) {
	r := newRollsum()
	data := []byte("Hello, World!")

	n, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	if r.blobSize != int64(len(data)) {
		t.Errorf("Expected blobSize %d, got %d", len(data), r.blobSize)
	}
}

func TestRollsumShouldSplit(t *testing.T) {
	r := newRollsum()

	// Split when blob reaches maximum size
	r.blobSize = maxBlobSize
	if !r.shouldSplit() {
		t.Error("Expected shouldSplit to return true for maxBlobSize")
	}

	// Split when rollsum indicates a good boundary
	r.blobSize = 100
	r.s2 = bupBlobSize - 1
	if !r.shouldSplit() {
		t.Error("Expected shouldSplit to return true for rollsum condition")
	}

	// Don't split for normal conditions
	r.blobSize = 100
	r.s2 = 12345
	if r.shouldSplit() {
		t.Error("Expected shouldSplit to return false")
	}
}

func TestRollsumGetBlobs(t *testing.T) {
	r := newRollsum()
	data := []byte("test data for blob creation")

	_, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	blobs := r.GetBlobs()

	if len(blobs) == 0 {
		t.Error("Expected at least one blob")
	}

	totalSize := int64(0)
	for _, blob := range blobs {
		totalSize += blob.size
	}

	if totalSize != int64(len(data)) {
		t.Errorf("Expected total blob size %d, got %d", len(data), totalSize)
	}
}

func TestComputeRollsumMatches(t *testing.T) {
	from := []rollsumBlob{
		{offset: 0, size: 10, crc32: 123},
		{offset: 10, size: 5, crc32: 456},
		{offset: 15, size: 8, crc32: 789},
	}

	to := []rollsumBlob{
		{offset: 0, size: 10, crc32: 123}, // Match with first from blob
		{offset: 10, size: 7, crc32: 456}, // Different size, no match
		{offset: 17, size: 8, crc32: 789}, // Match with third from blob
		{offset: 25, size: 3, crc32: 999}, // No match
	}

	matches := computeRollsumMatches(from, to)

	if len(matches.matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches.matches))
	}

	if matches.matchRatio != 50 { // 2 matches out of 4 to blobs = 50%
		t.Errorf("Expected match ratio 50, got %d", matches.matchRatio)
	}
}

func TestComputeRollsumMatchesEmptyTo(t *testing.T) {
	// Test that computeRollsumMatches doesn't panic when 'to' slice is empty
	// (rollsum.go line 205 computes nMatches * 100 / len(to), which would panic if len(to) == 0)
	from := []rollsumBlob{
		{offset: 0, size: 10, crc32: 123},
	}

	to := []rollsumBlob{}

	matches := computeRollsumMatches(from, to)

	if len(matches.matches) != 0 {
		t.Errorf("Expected 0 matches for empty 'to', got %d", len(matches.matches))
	}

	// matchRatio should be 0 for empty 'to' (or panic was avoided)
	if matches.matchRatio != 0 {
		t.Errorf("Expected match ratio 0 for empty 'to', got %d", matches.matchRatio)
	}
}

func TestRollsumCrcCalculation(t *testing.T) {
	r := newRollsum()
	data := []byte("test")

	_, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	blobs := r.GetBlobs()

	if len(blobs) != 1 {
		t.Fatalf("Expected 1 blob, got %d", len(blobs))
	}

	expectedCrc := crc32.ChecksumIEEE(data)
	if blobs[0].crc32 != expectedCrc {
		t.Errorf("Expected CRC %d, got %d", expectedCrc, blobs[0].crc32)
	}
}

func TestRollsumWithLargeData(t *testing.T) {
	r := newRollsum()

	// Use modest size for testing blob splitting behavior
	testDataSize := maxBlobSize + 1000

	data := make([]byte, testDataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	blobs := r.GetBlobs()

	// Should have at least 2 blobs due to size limit
	if len(blobs) < 2 {
		t.Errorf("Expected at least 2 blobs, got %d", len(blobs))
	}

	// Verify no blob exceeds maxBlobSize
	for i, blob := range blobs {
		if blob.size > maxBlobSize {
			t.Errorf("Blob %d size %d exceeds maxBlobSize %d", i, blob.size, maxBlobSize)
		}
	}
}

func TestMakeCrcMap(t *testing.T) {
	blobs := []rollsumBlob{
		{offset: 0, size: 10, crc32: 123},
		{offset: 10, size: 5, crc32: 456},
		{offset: 15, size: 8, crc32: 123}, // Same CRC as first
	}

	crcMap := makeCrcMap(blobs)

	if len(crcMap) != 2 {
		t.Errorf("Expected 2 unique CRCs, got %d", len(crcMap))
	}

	if len(crcMap[123]) != 2 {
		t.Errorf("Expected 2 blobs with CRC 123, got %d", len(crcMap[123]))
	}

	if len(crcMap[456]) != 1 {
		t.Errorf("Expected 1 blob with CRC 456, got %d", len(crcMap[456]))
	}
}

func TestRollsumFlush(t *testing.T) {
	r := newRollsum()
	// Set blobSize without writing data to test the flush bookkeeping mechanism
	// (we're not testing CRC calculation here, just that flush creates a new blob entry)
	r.blobSize = 50

	originalBlobCount := len(r.blobs)
	r.flush()

	if len(r.blobs) != originalBlobCount+1 {
		t.Errorf("Expected %d blobs after flush, got %d", originalBlobCount+1, len(r.blobs))
	}

	// Test flush with no pending blob
	r.flush()
	if len(r.blobs) != originalBlobCount+1 {
		t.Errorf("Expected %d blobs after second flush, got %d", originalBlobCount+1, len(r.blobs))
	}
}

func TestRollsumGetHeader(t *testing.T) {
	r := newRollsum()
	data := []byte("Hello, World! This is test data.")

	_, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	header := r.GetHeader()

	expectedHeader := data
	if len(data) > cap(r.header) {
		expectedHeader = data[:cap(r.header)]
	}

	if len(header) != len(expectedHeader) {
		t.Errorf("Expected header length %d, got %d", len(expectedHeader), len(header))
	}
}
