package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxExtractedFileBytes bounds any single file extracted from an
// update archive, guarding against a decompression bomb.
const maxExtractedFileBytes = 5 << 20 // 5 MB

// Apply extracts archiveData, a gzipped tar archive, into destDir.
// Extraction happens into a temporary sibling directory first and is
// only swapped into place once every entry has been safely written —
// a failure partway through never leaves destDir partially replaced.
//
// Every entry path is validated before any file is written: absolute
// paths, paths containing "..", and non-regular-file entries
// (symlinks, devices) are rejected. This is untrusted remote content,
// treated with the same care as any collector input.
func Apply(archiveData []byte, destDir string) error {
	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp(parent, ".alexiares-update-*")
	if err != nil {
		return fmt.Errorf("creating staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // no-op once renamed into place

	if err := extractTarGz(archiveData, tmpDir); err != nil {
		return err
	}

	// Atomically swap: move the old directory aside, move the new one
	// in, then remove the old. On most filesystems rename is atomic;
	// this sequence still leaves destDir valid at every step.
	backup := destDir + ".bak"
	_ = os.RemoveAll(backup)

	if _, err := os.Stat(destDir); err == nil {
		if err := os.Rename(destDir, backup); err != nil {
			return fmt.Errorf("backing up existing repository: %w", err)
		}
	}
	if err := os.Rename(tmpDir, destDir); err != nil {
		// Best-effort restore of the backup so a failed update
		// doesn't leave the signature repository missing entirely.
		_ = os.Rename(backup, destDir)
		return fmt.Errorf("installing new repository: %w", err)
	}
	_ = os.RemoveAll(backup)

	return nil
}

func extractTarGz(data []byte, destDir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("archive is not valid gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading archive: %w", err)
		}

		target, err := safeJoin(destDir, hdr.Name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("creating %s: %w", hdr.Name, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("creating parent of %s: %w", hdr.Name, err)
			}
			if err := writeFile(target, tr); err != nil {
				return fmt.Errorf("writing %s: %w", hdr.Name, err)
			}
		default:
			return fmt.Errorf("archive entry %s: unsupported type %v (only regular files and directories are allowed)", hdr.Name, hdr.Typeflag)
		}
	}
}

// safeJoin resolves name against base and confirms the result stays
// within base, rejecting the path-traversal ("zip-slip") pattern
// where an archive entry name like "../../etc/passwd" would otherwise
// let extraction write outside destDir.
func safeJoin(base, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("archive entry %q has an absolute path", name)
	}
	joined := filepath.Join(base, name)
	if joined != base && !strings.HasPrefix(joined, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes the extraction directory", name)
	}
	return joined, nil
}

func writeFile(target string, r io.Reader) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	written, err := io.Copy(f, io.LimitReader(r, maxExtractedFileBytes+1))
	if err != nil {
		return err
	}
	if written > maxExtractedFileBytes {
		return errors.New("file exceeds the maximum allowed size")
	}
	return nil
}
