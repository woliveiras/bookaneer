package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woliveiras/bookaneer/internal/database"
)

func TestBuildWhereClause(t *testing.T) {
	tests := []struct {
		name       string
		conditions []string
		want       string
	}{
		{"empty", nil, ""},
		{"empty slice", []string{}, ""},
		{"single", []string{"id = ?"}, "WHERE id = ?"},
		{"multiple", []string{"a = ?", "b = ?", "c = ?"}, "WHERE a = ? AND b = ? AND c = ?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, database.BuildWhereClause(tt.conditions))
		})
	}
}

func TestNormaliseSortDir(t *testing.T) {
	assert.Equal(t, "DESC", database.NormaliseSortDir("desc"))
	assert.Equal(t, "ASC", database.NormaliseSortDir("asc"))
	assert.Equal(t, "ASC", database.NormaliseSortDir(""))
	assert.Equal(t, "ASC", database.NormaliseSortDir("DESC")) // only "desc" maps to DESC
}

func TestClampLimit(t *testing.T) {
	assert.Equal(t, 50, database.ClampLimit(0, 50, 500))   // zero → default
	assert.Equal(t, 50, database.ClampLimit(-1, 50, 500))  // negative → default
	assert.Equal(t, 50, database.ClampLimit(501, 50, 500)) // over max → default
	assert.Equal(t, 25, database.ClampLimit(25, 50, 500))  // valid → used
	assert.Equal(t, 500, database.ClampLimit(500, 50, 500)) // at max → valid
	assert.Equal(t, 1, database.ClampLimit(1, 50, 500))    // min valid
}
