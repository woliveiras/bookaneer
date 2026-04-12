package pathmapping

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"/data/downloads", "/data/downloads/"},
		{"/data/downloads/", "/data/downloads/"},
		{"/data/downloads//", "/data/downloads/"},
		{"/media/library", "/media/library/"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := normalizePath(tt.give)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
