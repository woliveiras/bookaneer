package mediafile

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractMetadata_UnsupportedFormat(t *testing.T) {
	_, err := ExtractMetadata("test.pdf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestExtractMetadata_InvalidFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "invalid.epub")
	require.NoError(t, os.WriteFile(tmp, []byte("not a zip"), 0644))

	_, err := ExtractMetadata(tmp)
	assert.Error(t, err)
}

func TestExtractMetadata_ValidEPUB(t *testing.T) {
	epub := createTestEPUB(t,
		"META-INF/container.xml",
		`<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
		"OEBPS/content.opf",
		`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Dom Casmurro</dc:title>
    <dc:creator>Machado de Assis</dc:creator>
    <dc:language>pt</dc:language>
    <dc:publisher>Editora Garnier</dc:publisher>
    <dc:identifier opf:scheme="ISBN">9788525432230</dc:identifier>
    <dc:date>1899</dc:date>
    <dc:subject>Fiction</dc:subject>
    <dc:subject>Romance</dc:subject>
    <dc:description>A novel about jealousy and doubt.</dc:description>
  </metadata>
</package>`,
	)

	meta, err := ExtractMetadata(epub)
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, "Dom Casmurro", meta.Title)
	assert.Equal(t, []string{"Machado de Assis"}, meta.Authors)
	assert.Equal(t, "pt", meta.Language)
	assert.Equal(t, "Editora Garnier", meta.Publisher)
	assert.Equal(t, "9788525432230", meta.ISBN)
	assert.Equal(t, "1899", meta.Date)
	assert.Equal(t, []string{"Fiction", "Romance"}, meta.Subject)
	assert.Equal(t, "A novel about jealousy and doubt.", meta.Description)
}

func TestExtractMetadata_MinimalEPUB(t *testing.T) {
	epub := createTestEPUB(t,
		"META-INF/container.xml",
		`<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
		"content.opf",
		`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
  </metadata>
</package>`,
	)

	meta, err := ExtractMetadata(epub)
	require.NoError(t, err)
	assert.Equal(t, "Test Book", meta.Title)
	assert.Empty(t, meta.Authors)
	assert.Empty(t, meta.ISBN)
}

// createTestEPUB creates a minimal EPUB (zip) file with the given entries.
// Files are specified as alternating name/content pairs.
func createTestEPUB(t *testing.T, nameContentPairs ...string) string {
	t.Helper()
	if len(nameContentPairs)%2 != 0 {
		t.Fatal("nameContentPairs must have even number of elements")
	}

	path := filepath.Join(t.TempDir(), "test.epub")
	f, err := os.Create(path)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	for i := 0; i < len(nameContentPairs); i += 2 {
		fw, err := w.Create(nameContentPairs[i])
		require.NoError(t, err)
		_, err = fw.Write([]byte(nameContentPairs[i+1]))
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
	return path
}
