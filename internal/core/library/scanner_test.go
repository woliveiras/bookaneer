package library

import "testing"

func TestParsePathForBookInfo(t *testing.T) {
	tests := []struct {
		name       string
		rootPath   string
		filePath   string
		wantAuthor string
		wantTitle  string
	}{
		{
			name:       "author dir and book file",
			rootPath:   "/library",
			filePath:   "/library/J.R.R. Tolkien/The Hobbit.epub",
			wantAuthor: "J.R.R. Tolkien",
			wantTitle:  "The Hobbit",
		},
		{
			name:       "author dir, book dir, and book file",
			rootPath:   "/library",
			filePath:   "/library/J.R.R. Tolkien/The Hobbit/The Hobbit.epub",
			wantAuthor: "J.R.R. Tolkien",
			wantTitle:  "The Hobbit",
		},
		{
			name:       "flat layout with author-title separator",
			rootPath:   "/library",
			filePath:   "/library/J.R.R. Tolkien - The Hobbit.epub",
			wantAuthor: "J.R.R. Tolkien",
			wantTitle:  "The Hobbit",
		},
		{
			name:       "flat layout without author info",
			rootPath:   "/library",
			filePath:   "/library/The Hobbit.epub",
			wantAuthor: "",
			wantTitle:  "The Hobbit",
		},
		{
			name:       "pdf format",
			rootPath:   "/books",
			filePath:   "/books/Stephen King/It.pdf",
			wantAuthor: "Stephen King",
			wantTitle:  "It",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuthor, gotTitle := parsePathForBookInfo(tt.rootPath, tt.filePath)
			if gotAuthor != tt.wantAuthor {
				t.Errorf("author: got %q, want %q", gotAuthor, tt.wantAuthor)
			}
			if gotTitle != tt.wantTitle {
				t.Errorf("title: got %q, want %q", gotTitle, tt.wantTitle)
			}
		})
	}
}

func TestMakeSortName(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"J.R.R. Tolkien", "Tolkien, J.R.R."},
		{"Stephen King", "King, Stephen"},
		{"Tolkien", "Tolkien"},
		{"George R. R. Martin", "Martin, George R. R."},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := makeSortName(tt.give)
			if got != tt.want {
				t.Errorf("makeSortName(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
