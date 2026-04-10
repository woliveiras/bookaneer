package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	e := New(nil)

	tests := []struct {
		name     string
		settings Settings
		ctx      Context
		wantDir  string
		wantFile string
		wantRel  string
	}{
		{
			name:     "default template",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Author - $Title"},
			ctx:      Context{Author: "Machado de Assis", Title: "Dom Casmurro", Format: "epub"},
			wantDir:  "Machado de Assis",
			wantFile: "Machado de Assis - Dom Casmurro.epub",
			wantRel:  "Machado de Assis/Machado de Assis - Dom Casmurro.epub",
		},
		{
			name:     "sort author folder",
			settings: Settings{AuthorFolderFormat: "$SortAuthor", BookFileFormat: "$Title"},
			ctx:      Context{Author: "Machado de Assis", SortAuthor: "Assis, Machado de", Title: "Dom Casmurro", Format: "pdf"},
			wantDir:  "Assis, Machado de",
			wantFile: "Dom Casmurro.pdf",
			wantRel:  "Assis, Machado de/Dom Casmurro.pdf",
		},
		{
			name:     "with series conditional included",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Author - $Title{ ($SeriesName #$SeriesPosition)}"},
			ctx:      Context{Author: "Brandon Sanderson", Title: "The Way of Kings", Series: "Stormlight Archive", SeriesPosition: "1", Format: "epub"},
			wantDir:  "Brandon Sanderson",
			wantFile: "Brandon Sanderson - The Way of Kings (Stormlight Archive #1).epub",
			wantRel:  "Brandon Sanderson/Brandon Sanderson - The Way of Kings (Stormlight Archive #1).epub",
		},
		{
			name:     "series conditional omitted when no series",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Author - $Title{ ($SeriesName #$SeriesPosition)}"},
			ctx:      Context{Author: "Machado de Assis", Title: "Dom Casmurro", Format: "epub"},
			wantDir:  "Machado de Assis",
			wantFile: "Machado de Assis - Dom Casmurro.epub",
			wantRel:  "Machado de Assis/Machado de Assis - Dom Casmurro.epub",
		},
		{
			name:     "series conditional omitted when only series name",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Author - $Title{ ($SeriesName #$SeriesPosition)}"},
			ctx:      Context{Author: "Author", Title: "Title", Series: "MySeries", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Author - Title.epub",
			wantRel:  "Author/Author - Title.epub",
		},
		{
			name:     "replace spaces with dots",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title", ReplaceSpaces: true},
			ctx:      Context{Author: "Stephen King", Title: "The Shining", Format: "epub"},
			wantDir:  "Stephen.King",
			wantFile: "The.Shining.epub",
			wantRel:  "Stephen.King/The.Shining.epub",
		},
		{
			name:     "colon replacement dash",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title", ColonReplacement: "dash"},
			ctx:      Context{Author: "Author", Title: "Book: A Subtitle", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Book - A Subtitle.epub",
			wantRel:  "Author/Book - A Subtitle.epub",
		},
		{
			name:     "colon replacement delete",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title", ColonReplacement: "delete"},
			ctx:      Context{Author: "Author", Title: "Book: A Subtitle", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Book A Subtitle.epub",
			wantRel:  "Author/Book A Subtitle.epub",
		},
		{
			name:     "colon replacement space",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title", ColonReplacement: "space"},
			ctx:      Context{Author: "Author", Title: "Book: Subtitle", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Book Subtitle.epub",
			wantRel:  "Author/Book Subtitle.epub",
		},
		{
			name:     "unsafe characters sanitized in filename",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title"},
			ctx:      Context{Author: "Author", Title: "What? *Really*", Format: "pdf"},
			wantDir:  "Author",
			wantFile: "What_ _Really_.pdf",
			wantRel:  "Author/What_ _Really_.pdf",
		},
		{
			name:     "format extension not duplicated",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title.$Format"},
			ctx:      Context{Author: "Author", Title: "Book", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Book.epub",
			wantRel:  "Author/Book.epub",
		},
		{
			name:     "empty format produces no extension",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title"},
			ctx:      Context{Author: "Author", Title: "Book"},
			wantDir:  "Author",
			wantFile: "Book",
			wantRel:  "Author/Book",
		},
		{
			name:     "year variable",
			settings: Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title ($Year)"},
			ctx:      Context{Author: "Author", Title: "Book", Year: "2024", Format: "epub"},
			wantDir:  "Author",
			wantFile: "Book (2024).epub",
			wantRel:  "Author/Book (2024).epub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.Format("/library", tt.ctx, &tt.settings)
			assert.Equal(t, tt.wantDir, result.AuthorFolder)
			assert.Equal(t, tt.wantFile, result.Filename)
			assert.Equal(t, tt.wantRel, result.RelativePath)
			assert.Equal(t, "/library/"+tt.wantRel, result.FullPath)
		})
	}
}

func TestPreviewIsSameAsFormat(t *testing.T) {
	e := New(nil)
	s := &Settings{AuthorFolderFormat: "$Author", BookFileFormat: "$Title"}
	nc := Context{Author: "Author", Title: "Book", Format: "epub"}

	format := e.Format("/root", nc, s)
	preview := e.Preview("/root", nc, s)

	assert.Equal(t, format, preview)
}

func TestFormatDefaultSettings(t *testing.T) {
	e := New(nil)
	nc := Context{Author: "Author", Title: "Book", Format: "epub"}
	result := e.Format("/root", nc, nil)

	assert.Equal(t, "Author", result.AuthorFolder)
	assert.Equal(t, "Author - Book.epub", result.Filename)
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"normal.epub", "normal.epub"},
		{"what?really", "what_really"},
		{"a/b\\c*d", "a_b_c_d"},
		{"  spaced  ", "spaced"},
		{".hidden", "hidden"},
		{"trailing.", "trailing"},
		{"a<b>c|d\"e", "a_b_c_d_e"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeFilename(tt.give))
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"Author/Book", "Author/Book"},
		{"Author\\Book", "Author_Book"},
		{"a*b?c", "a_b_c"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizePath(tt.give))
		})
	}
}

func TestIsValidTemplate(t *testing.T) {
	tests := []struct {
		tmpl string
		want bool
	}{
		{"$Author - $Title", true},
		{"$Author/$Title ($Year)", true},
		{"$Author - $Title{ ($SeriesName #$SeriesPosition)}", true},
		{"$Unknown variable", false},
		{"plain text", true},
	}
	for _, tt := range tests {
		t.Run(tt.tmpl, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTemplate(tt.tmpl))
		})
	}
}

func TestCollapseRepeated(t *testing.T) {
	assert.Equal(t, "a b c", collapseRepeated("a   b  c", ' '))
	assert.Equal(t, "a.b.c", collapseRepeated("a...b..c", '.'))
}

func TestSampleContext(t *testing.T) {
	sc := SampleContext()
	require.NotEmpty(t, sc.Author)
	require.NotEmpty(t, sc.Title)
	require.NotEmpty(t, sc.Format)
}

func TestProcessConditionals(t *testing.T) {
	nc := Context{Author: "A", Title: "T", Series: "S", SeriesPosition: "1"}
	result := processConditionals("$Title{ - $Series}", nc)
	assert.Equal(t, "$Title - S", result) // $Series substituted inside the block

	nc2 := Context{Author: "A", Title: "T"}
	result2 := processConditionals("$Title{ - $Series}", nc2)
	assert.Equal(t, "$Title", result2)
}

func TestReplaceColons(t *testing.T) {
	assert.Equal(t, "A - B", replaceColons("A: B", "dash"))
	assert.Equal(t, "A  B", replaceColons("A: B", "space"))
	assert.Equal(t, "A B", replaceColons("A: B", "delete"))
	assert.Equal(t, "A - B", replaceColons("A: B", "unknown"))
}
