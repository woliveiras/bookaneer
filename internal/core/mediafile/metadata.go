package mediafile

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Metadata holds extracted information from an ebook file.
type Metadata struct {
	Title       string   `json:"title,omitempty"`
	Authors     []string `json:"authors,omitempty"`
	Publisher   string   `json:"publisher,omitempty"`
	Language    string   `json:"language,omitempty"`
	Description string   `json:"description,omitempty"`
	ISBN        string   `json:"isbn,omitempty"`
	Subject     []string `json:"subject,omitempty"`
	Date        string   `json:"date,omitempty"`
}

// ExtractMetadata reads metadata from an ebook file.
// Supported formats: EPUB.
func ExtractMetadata(path string) (*Metadata, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".epub":
		return extractEPUBMetadata(path)
	default:
		return nil, fmt.Errorf("unsupported format for metadata extraction: %s", ext)
	}
}

// epubContainer represents the META-INF/container.xml file.
type epubContainer struct {
	XMLName   xml.Name       `xml:"container"`
	RootFiles []epubRootFile `xml:"rootfiles>rootfile"`
}

type epubRootFile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

// opfPackage represents the OPF package document.
type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
}

type opfMetadata struct {
	Title       []string        `xml:"title"`
	Creator     []opfCreator    `xml:"creator"`
	Publisher   []string        `xml:"publisher"`
	Language    []string        `xml:"language"`
	Description []string        `xml:"description"`
	Subject     []string        `xml:"subject"`
	Date        []string        `xml:"date"`
	Identifier  []opfIdentifier `xml:"identifier"`
}

type opfCreator struct {
	Value string `xml:",chardata"`
	Role  string `xml:"role,attr"`
}

type opfIdentifier struct {
	Value  string `xml:",chardata"`
	Scheme string `xml:"scheme,attr"`
}

func extractEPUBMetadata(path string) (*Metadata, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}
	defer r.Close()

	// Find OPF path from container.xml
	opfPath, err := findOPFPath(&r.Reader)
	if err != nil {
		return nil, fmt.Errorf("find opf: %w", err)
	}

	// Read OPF file
	opfFile, err := findFile(&r.Reader, opfPath)
	if err != nil {
		return nil, fmt.Errorf("open opf: %w", err)
	}

	rc, err := opfFile.Open()
	if err != nil {
		return nil, fmt.Errorf("read opf: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("read opf data: %w", err)
	}

	var pkg opfPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse opf: %w", err)
	}

	meta := &Metadata{}
	m := pkg.Metadata

	if len(m.Title) > 0 {
		meta.Title = m.Title[0]
	}
	for _, c := range m.Creator {
		if c.Value != "" {
			meta.Authors = append(meta.Authors, c.Value)
		}
	}
	if len(m.Publisher) > 0 {
		meta.Publisher = m.Publisher[0]
	}
	if len(m.Language) > 0 {
		meta.Language = m.Language[0]
	}
	if len(m.Description) > 0 {
		meta.Description = m.Description[0]
	}
	if len(m.Date) > 0 {
		meta.Date = m.Date[0]
	}
	meta.Subject = m.Subject

	// Extract ISBN from identifiers
	for _, id := range m.Identifier {
		scheme := strings.ToLower(id.Scheme)
		val := strings.TrimSpace(id.Value)
		if scheme == "isbn" || strings.HasPrefix(val, "978") || strings.HasPrefix(val, "979") {
			// Clean ISBN
			cleaned := strings.ReplaceAll(val, "-", "")
			cleaned = strings.ReplaceAll(cleaned, " ", "")
			if len(cleaned) == 10 || len(cleaned) == 13 {
				meta.ISBN = cleaned
				break
			}
		}
	}

	return meta, nil
}

func findOPFPath(r *zip.Reader) (string, error) {
	containerFile, err := findFile(r, "META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("container.xml not found: %w", err)
	}

	rc, err := containerFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 1<<20))
	if err != nil {
		return "", err
	}

	var container epubContainer
	if err := xml.Unmarshal(data, &container); err != nil {
		return "", fmt.Errorf("parse container.xml: %w", err)
	}

	for _, rf := range container.RootFiles {
		if rf.MediaType == "application/oebps-package+xml" || strings.HasSuffix(rf.FullPath, ".opf") {
			return rf.FullPath, nil
		}
	}

	return "", fmt.Errorf("no OPF rootfile found in container.xml")
}

func findFile(r *zip.Reader, name string) (*zip.File, error) {
	for _, f := range r.File {
		if f.Name == name || strings.EqualFold(f.Name, name) {
			return f, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}
