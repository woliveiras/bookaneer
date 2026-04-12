package mediafile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyContent(t *testing.T) {
	tests := []struct {
		name           string
		meta           *Metadata
		expectedTitle  string
		expectedAuthor string
		wantMismatch   bool
	}{
		{
			name:           "exact match",
			meta:           &Metadata{Title: "Dom Casmurro", Authors: []string{"Machado de Assis"}},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false,
		},
		{
			name:           "complete mismatch",
			meta:           &Metadata{Title: "Harry Potter and the Sorcerer's Stone", Authors: []string{"J.K. Rowling"}},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   true,
		},
		{
			name:           "same author different book",
			meta:           &Metadata{Title: "Quincas Borba", Authors: []string{"Machado de Assis"}},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false, // author matches, just different book/edition
		},
		{
			name:           "same title different author format",
			meta:           &Metadata{Title: "Dom Casmurro", Authors: []string{"Assis, Machado de"}},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false, // title matches perfectly
		},
		{
			name:           "nil metadata",
			meta:           nil,
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false, // can't verify, assume OK
		},
		{
			name:           "empty metadata",
			meta:           &Metadata{},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false, // no data to compare, assume OK
		},
		{
			name:           "subtitle difference",
			meta:           &Metadata{Title: "Dom Casmurro: A Novel", Authors: []string{"Machado de Assis"}},
			expectedTitle:  "Dom Casmurro",
			expectedAuthor: "Machado de Assis",
			wantMismatch:   false,
		},
		{
			name:           "completely unrelated content",
			meta:           &Metadata{Title: "Cooking Recipes for Beginners", Authors: []string{"John Smith"}},
			expectedTitle:  "Dune",
			expectedAuthor: "Frank Herbert",
			wantMismatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyContent(tt.meta, tt.expectedTitle, tt.expectedAuthor)
			assert.Equal(t, tt.wantMismatch, result.Mismatch, "mismatch flag: title=%.2f, author=%.2f", result.TitleScore, result.AuthorScore)
		})
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		min  float64
		max  float64
	}{
		{"identical", "dom casmurro", "dom casmurro", 1.0, 1.0},
		{"similar", "dom casmurro", "dom casmurro a novel", 0.5, 1.0},
		{"different", "dom casmurro", "harry potter", 0.0, 0.2},
		{"empty", "", "", 1.0, 1.0},
		{"one empty", "hello", "", 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := similarity(tt.a, tt.b)
			assert.GreaterOrEqual(t, score, tt.min, "score %.2f below min %.2f", score, tt.min)
			assert.LessOrEqual(t, score, tt.max, "score %.2f above max %.2f", score, tt.max)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"Dom Casmurro", "dom casmurro"},
		{"J.K. Rowling", "jk rowling"},
		{"  spaces  ", "spaces"},
		{"L'Étranger", "létranger"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, normalize(tt.give))
		})
	}
}
