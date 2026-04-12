package mediafile

import (
	"strings"
	"unicode"
)

// ContentMatch describes how well a file's metadata matches expected values.
type ContentMatch struct {
	TitleScore  float64 `json:"titleScore"`
	AuthorScore float64 `json:"authorScore"`
	Mismatch    bool    `json:"mismatch"`
}

// MismatchThreshold is the minimum similarity score to consider content matching.
// Below this value, the content is considered a mismatch.
const MismatchThreshold = 0.35

// VerifyContent compares extracted metadata against expected book title and author.
// Returns a ContentMatch indicating whether the file content matches expectations.
func VerifyContent(meta *Metadata, expectedTitle, expectedAuthor string) ContentMatch {
	if meta == nil {
		return ContentMatch{Mismatch: false} // Can't verify, assume OK
	}

	result := ContentMatch{}

	// Compare title
	if meta.Title != "" && expectedTitle != "" {
		result.TitleScore = similarity(normalize(meta.Title), normalize(expectedTitle))
	} else {
		result.TitleScore = 1.0 // No data to compare, assume OK
	}

	// Compare author — check if any extracted author matches any expected author word
	if len(meta.Authors) > 0 && expectedAuthor != "" {
		result.AuthorScore = bestAuthorMatch(meta.Authors, expectedAuthor)
	} else {
		result.AuthorScore = 1.0 // No data to compare, assume OK
	}

	// Content is a mismatch if BOTH title AND author are below threshold.
	// If title is very different but author matches, it could be a different edition.
	// If author is different but title matches, could be a translation or collection.
	result.Mismatch = result.TitleScore < MismatchThreshold && result.AuthorScore < MismatchThreshold

	return result
}

// bestAuthorMatch finds the best similarity between extracted authors and the expected author.
func bestAuthorMatch(extractedAuthors []string, expectedAuthor string) float64 {
	normalizedExpected := normalize(expectedAuthor)
	best := 0.0
	for _, author := range extractedAuthors {
		score := similarity(normalize(author), normalizedExpected)
		if score > best {
			best = score
		}
		// Also check if last names match (common in publishing metadata)
		if lastNameMatch(author, expectedAuthor) {
			if score < 0.5 {
				score = 0.5
			}
			if score > best {
				best = score
			}
		}
	}
	return best
}

// lastNameMatch checks if the last word (assumed to be the surname) matches.
func lastNameMatch(a, b string) bool {
	aWords := strings.Fields(normalize(a))
	bWords := strings.Fields(normalize(b))
	if len(aWords) == 0 || len(bWords) == 0 {
		return false
	}
	return aWords[len(aWords)-1] == bWords[len(bWords)-1]
}

// normalize lowercases and removes non-alphanumeric characters (except spaces).
func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// similarity returns a score between 0 and 1 using bigram overlap (Dice coefficient).
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) < 2 || len(b) < 2 {
		if a == b {
			return 1.0
		}
		// For very short strings, check containment
		if strings.Contains(a, b) || strings.Contains(b, a) {
			return 0.8
		}
		return 0.0
	}

	aBigrams := bigrams(a)
	bBigrams := bigrams(b)

	intersection := 0
	for bg := range aBigrams {
		if count, ok := bBigrams[bg]; ok {
			if aBigrams[bg] < count {
				intersection += aBigrams[bg]
			} else {
				intersection += count
			}
		}
	}

	return 2.0 * float64(intersection) / float64(len(a)-1+len(b)-1)
}

// bigrams returns a map of character bigrams and their counts.
func bigrams(s string) map[string]int {
	runes := []rune(s)
	m := make(map[string]int, len(runes)-1)
	for i := 0; i < len(runes)-1; i++ {
		bg := string(runes[i : i+2])
		m[bg]++
	}
	return m
}
