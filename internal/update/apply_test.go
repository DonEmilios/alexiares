package update_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexiares/alexiares/internal/update"
)

// buildTarGz constructs an in-memory gzipped tar archive from the
// given entries (path -> file content). A path ending in "/" is
// written as a directory entry.
func buildTarGz(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, content := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if name[len(name)-1] == '/' {
			hdr.Typeflag = tar.TypeDir
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader(%s) error = %v", name, err)
		}
		if hdr.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(content)); err != nil {
				t.Fatalf("Write(%s) error = %v", name, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar Close() error = %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close() error = %v", err)
	}
	return buf.Bytes()
}

func TestApplyExtractsFiles(t *testing.T) {
	archive := buildTarGz(t, map[string]string{
		"favicon/cluster001.yaml": "id: cluster001\n",
		"wallets/cluster002.yaml": "id: cluster002\n",
	})

	dest := filepath.Join(t.TempDir(), "signatures")
	if err := update.Apply(archive, dest); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "favicon", "cluster001.yaml"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(got) != "id: cluster001\n" {
		t.Errorf("extracted content = %q, want %q", got, "id: cluster001\n")
	}
}

func TestApplyRejectsPathTraversal(t *testing.T) {
	archive := buildTarGz(t, map[string]string{
		"../../etc/evil.yaml": "malicious",
	})

	dest := filepath.Join(t.TempDir(), "signatures")
	err := update.Apply(archive, dest)
	if err == nil {
		t.Fatal("Apply() error = nil, want rejection of a path-traversal entry")
	}

	// Confirm nothing escaped: the parent-of-parent directory must not
	// contain the malicious file.
	escaped := filepath.Join(filepath.Dir(filepath.Dir(dest)), "etc", "evil.yaml")
	if _, statErr := os.Stat(escaped); statErr == nil {
		t.Fatal("path traversal succeeded: file was written outside the extraction directory")
	}
}

func TestApplyRejectsAbsolutePath(t *testing.T) {
	archive := buildTarGz(t, map[string]string{
		"/etc/evil.yaml": "malicious",
	})
	dest := filepath.Join(t.TempDir(), "signatures")
	if err := update.Apply(archive, dest); err == nil {
		t.Fatal("Apply() error = nil, want rejection of an absolute path entry")
	}
}

func TestApplyReplacesExistingRepository(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "signatures")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.yaml"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	archive := buildTarGz(t, map[string]string{"new.yaml": "fresh"})
	if err := update.Apply(archive, dest); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "old.yaml")); !os.IsNotExist(err) {
		t.Error("old.yaml still exists after Apply(), want it replaced")
	}
	if _, err := os.Stat(filepath.Join(dest, "new.yaml")); err != nil {
		t.Errorf("new.yaml missing after Apply(): %v", err)
	}
	if _, err := os.Stat(dest + ".bak"); !os.IsNotExist(err) {
		t.Error("backup directory left behind after successful Apply()")
	}
}

func TestApplyRejectsMalformedArchive(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "signatures")
	if err := update.Apply([]byte("not a gzip archive"), dest); err == nil {
		t.Fatal("Apply() error = nil, want rejection of malformed archive data")
	}
}
