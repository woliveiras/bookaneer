package naming

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Engine formats file and folder paths using configurable naming templates.
type Engine struct {
	db *sql.DB
}

// New creates a new naming Engine.
func New(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// Context holds all the variables available for template substitution.
type Context struct {
	Author         string
	SortAuthor     string
	Title          string
	Series         string
	SeriesPosition string
	Year           string
	Format         string
	Quality        string
	OriginalName   string
}

// Settings holds the configurable naming options.
type Settings struct {
	Enabled            bool
	AuthorFolderFormat string
	BookFileFormat     string
	ReplaceSpaces      bool
	ColonReplacement   string // "dash", "space", "delete"
}

// Result holds the formatted path components.
type Result struct {
	AuthorFolder string
	Filename     string
	RelativePath string
	FullPath     string
}

var defaultSettings = Settings{
	Enabled:            true,
	AuthorFolderFormat: "$Author",
	BookFileFormat:     "$Author - $Title",
	ReplaceSpaces:      false,
	ColonReplacement:   "dash",
}

// LoadSettings reads naming settings from the config table.
func (e *Engine) LoadSettings(ctx context.Context) (*Settings, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT key, value FROM config WHERE key LIKE 'naming.%'
	`)
	if err != nil {
		return nil, fmt.Errorf("query naming settings: %w", err)
	}
	defer rows.Close()

	s := defaultSettings
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan naming setting: %w", err)
		}
		switch key {
		case "naming.enabled":
			s.Enabled = value == "1"
		case "naming.authorFolderFormat":
			s.AuthorFolderFormat = value
		case "naming.bookFileFormat":
			s.BookFileFormat = value
		case "naming.replaceSpaces":
			s.ReplaceSpaces = value == "1"
		case "naming.colonReplacement":
			s.ColonReplacement = value
		}
	}
	return &s, rows.Err()
}

// SaveSettings persists naming settings to the config table.
func (e *Engine) SaveSettings(ctx context.Context, s *Settings) error {
	boolStr := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}

	pairs := []struct{ key, value string }{
		{"naming.enabled", boolStr(s.Enabled)},
		{"naming.authorFolderFormat", s.AuthorFolderFormat},
		{"naming.bookFileFormat", s.BookFileFormat},
		{"naming.replaceSpaces", boolStr(s.ReplaceSpaces)},
		{"naming.colonReplacement", s.ColonReplacement},
	}

	for _, p := range pairs {
		_, err := e.db.ExecContext(ctx,
			`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`,
			p.key, p.value,
		)
		if err != nil {
			return fmt.Errorf("save naming setting %s: %w", p.key, err)
		}
	}
	return nil
}

// Format builds the full destination path for a book file.
func (e *Engine) Format(rootPath string, nc Context, s *Settings) Result {
	if s == nil {
		s = &defaultSettings
	}

	authorFolder := e.applyTemplate(s.AuthorFolderFormat, nc, s)
	authorFolder = sanitizePath(authorFolder)

	filename := e.applyTemplate(s.BookFileFormat, nc, s)
	filename = sanitizeFilename(filename)

	// Append format extension
	ext := strings.ToLower(nc.Format)
	if ext != "" && !strings.HasSuffix(filename, "."+ext) {
		filename = filename + "." + ext
	}

	relativePath := filepath.Join(authorFolder, filename)
	fullPath := filepath.Join(rootPath, relativePath)

	return Result{
		AuthorFolder: authorFolder,
		Filename:     filename,
		RelativePath: relativePath,
		FullPath:     fullPath,
	}
}

// Preview returns what the path would look like without actually renaming.
// It is identical to Format and exists for API clarity.
func (e *Engine) Preview(rootPath string, nc Context, s *Settings) Result {
	return e.Format(rootPath, nc, s)
}

// applyTemplate substitutes $Variables in a template string.
// Supports conditional blocks: { ($SeriesName #$SeriesPosition)} omitted if variable is empty.
func (e *Engine) applyTemplate(tmpl string, nc Context, s *Settings) string {
	result := processConditionals(tmpl, nc)
	result = substituteVars(result, nc)
	result = replaceColons(result, s.ColonReplacement)

	if s.ReplaceSpaces {
		result = strings.ReplaceAll(result, " ", ".")
	}

	result = collapseRepeated(result, ' ')
	result = collapseRepeated(result, '.')

	return strings.TrimSpace(result)
}

// processConditionals handles {conditional blocks} in templates.
// A block is included only if ALL $Variables inside it have non-empty values.
func processConditionals(tmpl string, nc Context) string {
	var buf strings.Builder
	i := 0
	for i < len(tmpl) {
		if tmpl[i] == '{' {
			end := strings.IndexByte(tmpl[i:], '}')
			if end == -1 {
				buf.WriteByte(tmpl[i])
				i++
				continue
			}
			block := tmpl[i+1 : i+end]
			if allVarsPresent(block, nc) {
				buf.WriteString(substituteVars(block, nc))
			}
			i += end + 1
		} else {
			buf.WriteByte(tmpl[i])
			i++
		}
	}
	return buf.String()
}

func allVarsPresent(block string, nc Context) bool {
	vars := []struct {
		token string
		value string
	}{
		{"$SeriesName", nc.Series},
		{"$Series", nc.Series},
		{"$SeriesPosition", nc.SeriesPosition},
		{"$Author", nc.Author},
		{"$SortAuthor", nc.SortAuthor},
		{"$Title", nc.Title},
		{"$Year", nc.Year},
		{"$Format", nc.Format},
		{"$Quality", nc.Quality},
		{"$OriginalName", nc.OriginalName},
	}
	for _, v := range vars {
		if strings.Contains(block, v.token) && v.value == "" {
			return false
		}
	}
	return true
}

func substituteVars(s string, nc Context) string {
	replacements := []struct {
		token string
		value string
	}{
		{"$SeriesPosition", nc.SeriesPosition},
		{"$SeriesName", nc.Series},
		{"$SortAuthor", nc.SortAuthor},
		{"$OriginalName", nc.OriginalName},
		{"$Author", nc.Author},
		{"$Series", nc.Series},
		{"$Title", nc.Title},
		{"$Year", nc.Year},
		{"$Format", nc.Format},
		{"$Quality", nc.Quality},
	}
	for _, r := range replacements {
		s = strings.ReplaceAll(s, r.token, r.value)
	}
	return s
}

func replaceColons(s, mode string) string {
	switch mode {
	case "dash":
		return strings.ReplaceAll(s, ":", " -")
	case "space":
		return strings.ReplaceAll(s, ":", " ")
	case "delete":
		return strings.ReplaceAll(s, ":", "")
	default:
		return strings.ReplaceAll(s, ":", " -")
	}
}

// sanitizeFilename removes characters that are invalid in file names.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	name = replacer.Replace(name)
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	return name
}

// sanitizePath removes characters invalid in directory names but allows path separators.
func sanitizePath(name string) string {
	replacer := strings.NewReplacer(
		"\\", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	name = replacer.Replace(name)
	return strings.TrimSpace(name)
}

func collapseRepeated(s string, ch byte) string {
	var buf strings.Builder
	prev := false
	for i := range len(s) {
		if s[i] == ch {
			if !prev {
				buf.WriteByte(ch)
			}
			prev = true
		} else {
			buf.WriteByte(s[i])
			prev = false
		}
	}
	return buf.String()
}

// SampleContext returns a sample NamingContext for preview display.
func SampleContext() Context {
	return Context{
		Author:         "Machado de Assis",
		SortAuthor:     "Assis, Machado de",
		Title:          "Dom Casmurro",
		Series:         "Realismo Brasileiro",
		SeriesPosition: "1",
		Year:           "1899",
		Format:         "epub",
		Quality:        "epub",
		OriginalName:   "dom_casmurro.epub",
	}
}

// IsValidTemplate checks whether a template string uses only known variables.
func IsValidTemplate(tmpl string) bool {
	known := []string{
		"$Author", "$SortAuthor", "$Title", "$Series", "$SeriesName",
		"$SeriesPosition", "$Year", "$Format", "$Quality", "$OriginalName",
	}
	work := tmpl
	for _, k := range known {
		work = strings.ReplaceAll(work, k, "")
	}
	for _, r := range work {
		_ = r
		if strings.Contains(work, "$") {
			return false
		}
	}
	return true
}

// RenameResult holds the results of a batch rename operation.
type RenameResult struct {
	Total   int           `json:"total"`
	Renamed int           `json:"renamed"`
	Skipped int           `json:"skipped"`
	Errors  []string      `json:"errors,omitempty"`
	Files   []RenamedFile `json:"files,omitempty"`
}

// RenamedFile represents a single file rename operation.
type RenamedFile struct {
	BookID  int64  `json:"bookId"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// PreviewRenameAll returns what all files would be renamed to without actually renaming.
func (e *Engine) PreviewRenameAll(ctx context.Context) (*RenameResult, error) {
	return e.renameAll(ctx, true)
}

// RenameAll renames all library files according to current naming settings.
func (e *Engine) RenameAll(ctx context.Context) (*RenameResult, error) {
	return e.renameAll(ctx, false)
}

func (e *Engine) renameAll(ctx context.Context, dryRun bool) (*RenameResult, error) {
	settings, err := e.LoadSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load naming settings: %w", err)
	}

	rows, err := e.db.QueryContext(ctx, `
		SELECT bf.id, bf.book_id, bf.path, bf.format,
		       b.title, COALESCE(b.release_date, ''),
		       a.name, COALESCE(a.sort_name, ''),
		       COALESCE(s.title, ''), COALESCE(sb.position, ''),
		       rf.path as root_path
		FROM book_files bf
		JOIN books b ON bf.book_id = b.id
		JOIN authors a ON b.author_id = a.id
		LEFT JOIN series_books sb ON sb.book_id = b.id
		LEFT JOIN series s ON sb.series_id = s.id
		JOIN root_folders rf ON bf.path LIKE rf.path || '%'
		ORDER BY bf.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query book files: %w", err)
	}

	type fileInfo struct {
		id       int64
		bookID   int64
		path     string
		rootPath string
		nc       Context
	}

	var files []fileInfo
	for rows.Next() {
		var fi fileInfo
		var format, title, releaseDate, authorName, sortAuthor, seriesTitle, seriesPos string
		if err := rows.Scan(
			&fi.id, &fi.bookID, &fi.path, &format,
			&title, &releaseDate,
			&authorName, &sortAuthor,
			&seriesTitle, &seriesPos,
			&fi.rootPath,
		); err != nil {
			slog.Warn("Failed to scan book file for rename", "error", err)
			continue
		}

		year := ""
		if len(releaseDate) >= 4 {
			year = releaseDate[:4]
		}

		fi.nc = Context{
			Author:         authorName,
			SortAuthor:     sortAuthor,
			Title:          title,
			Series:         seriesTitle,
			SeriesPosition: seriesPos,
			Year:           year,
			Format:         format,
			OriginalName:   filepath.Base(fi.path),
		}
		files = append(files, fi)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate book files: %w", err)
	}
	rows.Close()

	result := &RenameResult{Total: len(files)}

	for _, fi := range files {
		named := e.Format(fi.rootPath, fi.nc, settings)

		if fi.path == named.FullPath {
			result.Skipped++
			continue
		}

		result.Files = append(result.Files, RenamedFile{
			BookID:  fi.bookID,
			OldPath: fi.path,
			NewPath: named.FullPath,
		})

		if dryRun {
			result.Renamed++
			continue
		}

		// Create destination directory
		destDir := filepath.Dir(named.FullPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("create dir for file %d: %v", fi.id, err))
			continue
		}

		// Rename file on disk
		if err := os.Rename(fi.path, named.FullPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("rename file %d: %v", fi.id, err))
			continue
		}

		// Update database
		_, err := e.db.ExecContext(ctx, `
			UPDATE book_files SET path = ?, relative_path = ? WHERE id = ?
		`, named.FullPath, named.RelativePath, fi.id)
		if err != nil {
			// Try to rename back on DB error
			_ = os.Rename(named.FullPath, fi.path)
			result.Errors = append(result.Errors, fmt.Sprintf("update db for file %d: %v", fi.id, err))
			continue
		}

		result.Renamed++

		// Try to remove old empty directory
		oldDir := filepath.Dir(fi.path)
		_ = os.Remove(oldDir)
	}

	return result, nil
}
