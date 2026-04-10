package wanted

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// hashFile calculates SHA256 hash of a file.
func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sanitizeFilename removes unsafe characters from filename.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}

// detectEbookFormat infers the ebook format from a string (title, URL, etc.).
// Returns "epub", "pdf", "mobi", or "unknown".
func detectEbookFormat(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "epub"):
		return "epub"
	case strings.Contains(s, "pdf"):
		return "pdf"
	case strings.Contains(s, "mobi"):
		return "mobi"
	}
	return "unknown"
}

// isEbookFormat reports whether the string contains a known ebook format keyword.
func isEbookFormat(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "epub") ||
		strings.Contains(s, "pdf") ||
		strings.Contains(s, "mobi") ||
		strings.Contains(s, "ebook")
}

// buildAuthorSavePath returns the author directory and the full expected save path
// for a download given a download root dir, author name, and filename.
func buildAuthorSavePath(downloadDir, authorName, filename string) (authorDir, savePath string) {
	authorDir = filepath.Join(downloadDir, sanitizeFilename(authorName))
	savePath = filepath.Join(authorDir, filename)
	return
}
