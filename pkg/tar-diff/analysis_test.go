package tar_diff

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

// Helper function to create test tar files in memory (for hardlink tests)
func createTestTar(entries []tarEntry) (io.ReadSeeker, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, entry := range entries {
		hdr := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Size:     int64(len(entry.data)),
			Mode:     0644,
		}
		if entry.linkname != "" {
			hdr.Linkname = entry.linkname
		}
		if entry.mode != 0 {
			hdr.Mode = entry.mode
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if len(entry.data) > 0 {
			if _, err := tw.Write(entry.data); err != nil {
				return nil, err
			}
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

type tarEntry struct {
	name     string
	typeflag byte
	data     []byte
	linkname string
	mode     int64
}

func TestAnalyzeTar_Hardlinks(t *testing.T) {
	entries := []tarEntry{
		{name: "original.txt", typeflag: tar.TypeReg, data: []byte("content")},
		{name: "hardlink.txt", typeflag: tar.TypeLink, linkname: "original.txt"},
	}
	tarFile, err := createTestTar(entries)
	if err != nil {
		t.Fatalf("Failed to create test tar: %v", err)
	}

	info, err := analyzeTar(tarFile)
	if err != nil {
		t.Fatalf("analyzeTar failed: %v", err)
	}

	// Should have 1 file (the original)
	if len(info.files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(info.files))
	}

	// Should have 1 hardlink
	if len(info.hardlinks) != 1 {
		t.Errorf("Expected 1 hardlink, got %d", len(info.hardlinks))
	}

	// Check hardlink details
	hl := info.hardlinks[0]
	if hl.path != "hardlink.txt" {
		t.Errorf("Expected hardlink path 'hardlink.txt', got %q", hl.path)
	}
	if hl.linkname != "original.txt" {
		t.Errorf("Expected linkname 'original.txt', got %q", hl.linkname)
	}
	if hl.index != 1 {
		t.Errorf("Expected hardlink index 1, got %d", hl.index)
	}
}

func TestAnalyzeTar_MultipleHardlinks(t *testing.T) {
	entries := []tarEntry{
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("content")},
		{name: "link1.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
		{name: "link2.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
		{name: "link3.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
	}
	tarFile, err := createTestTar(entries)
	if err != nil {
		t.Fatalf("Failed to create test tar: %v", err)
	}

	info, err := analyzeTar(tarFile)
	if err != nil {
		t.Fatalf("analyzeTar failed: %v", err)
	}

	if len(info.files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(info.files))
	}
	if len(info.hardlinks) != 3 {
		t.Errorf("Expected 3 hardlinks, got %d", len(info.hardlinks))
	}

	// Verify all hardlinks point to the same file
	for i, hl := range info.hardlinks {
		if hl.linkname != "file.txt" {
			t.Errorf("Hardlink %d: expected linkname 'file.txt', got %q", i, hl.linkname)
		}
	}
}

func TestAnalyzeTar_DuplicateFilesAndHardlinks(t *testing.T) {
	entries := []tarEntry{
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("first")},
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("second")},
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("third")},
		{name: "link1.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
		{name: "link2.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
	}

	tarFile, err := createTestTar(entries)
	if err != nil {
		t.Fatalf("Failed to create test tar: %v", err)
	}

	info, err := analyzeTar(tarFile)
	if err != nil {
		t.Fatalf("analyzeTar failed: %v", err)
	}

	if len(info.files) != 3 {
		t.Errorf("Should have 3 files (all duplicates), got %d", len(info.files))
	}
	if len(info.hardlinks) != 2 {
		t.Errorf("Should have 2 hardlinks, got %d", len(info.hardlinks))
	}
}

func TestAnalyzeForDelta_HardlinksInTargetInfo(t *testing.T) {
	// Create old tar with a file
	oldEntries := []tarEntry{
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("old content")},
	}
	oldTar, err := createTestTar(oldEntries)
	if err != nil {
		t.Fatalf("Failed to create old tar: %v", err)
	}

	// Create new tar with the same file and a hardlink
	newEntries := []tarEntry{
		{name: "file.txt", typeflag: tar.TypeReg, data: []byte("old content")},
		{name: "link.txt", typeflag: tar.TypeLink, linkname: "file.txt"},
	}
	newTar, err := createTestTar(newEntries)
	if err != nil {
		t.Fatalf("Failed to create new tar: %v", err)
	}

	// Reset both tars
	if _, err := oldTar.Seek(0, 0); err != nil {
		t.Fatalf("oldTar.Seek: %v", err)
	}
	if _, err := newTar.Seek(0, 0); err != nil {
		t.Fatalf("newTar.Seek: %v", err)
	}

	oldInfo, err := analyzeTar(oldTar)
	if err != nil {
		t.Fatalf("analyzeTar (old) failed: %v", err)
	}

	newInfo, err := analyzeTar(newTar)
	if err != nil {
		t.Fatalf("analyzeTar (new) failed: %v", err)
	}

	// Reset old tar for analyzeForDelta
	if _, err := oldTar.Seek(0, 0); err != nil {
		t.Fatalf("oldTar.Seek: %v", err)
	}

	analysis, err := analyzeForDelta(oldInfo, newInfo, oldTar)
	if err != nil {
		t.Fatalf("analyzeForDelta failed: %v", err)
	}
	defer func() {
		if err := analysis.Close(); err != nil {
			t.Fatalf("analysis.Close failed: %v", err)
		}
	}()

	// Check that hardlink is in targetInfoByIndex
	hlInfo, exists := analysis.targetInfoByIndex[1]
	if !exists {
		t.Fatal("Hardlink should be found in targetInfoByIndex at index 1")
	}
	if hlInfo.hardlink == nil {
		t.Fatal("targetInfo.hardlink should not be nil")
	}
	if hlInfo.hardlink.path != "link.txt" {
		t.Errorf("Expected hardlink path 'link.txt', got %q", hlInfo.hardlink.path)
	}
	if hlInfo.hardlink.linkname != "file.txt" {
		t.Errorf("Expected linkname 'file.txt', got %q", hlInfo.hardlink.linkname)
	}
}
