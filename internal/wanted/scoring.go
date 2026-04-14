package wanted

import (
	"fmt"
	"strings"

	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// unifiedResult wraps either a library or indexer result with computed score.
type unifiedResult struct {
	// Source identification
	isLibrary   bool
	isIndexer   bool
	sourceName  string
	downloadURL string
	title       string
	format      string
	size        int64

	// Type-specific data (one will be nil)
	libraryResult *library.SearchResult
	indexerResult *search.Result

	// Computed score for ranking
	score int
}

// newLibraryResult creates a unified result from a library search result.
func newLibraryResult(r *library.SearchResult) *unifiedResult {
	return &unifiedResult{
		isLibrary:     true,
		sourceName:    r.Provider,
		downloadURL:   r.DownloadURL,
		title:         r.Title,
		format:        r.Format,
		size:          r.Size,
		libraryResult: r,
	}
}

// newIndexerResult creates a unified result from an indexer search result.
func newIndexerResult(r *search.Result) *unifiedResult {
	format := extractFormat(r.Title)
	return &unifiedResult{
		isIndexer:     true,
		sourceName:    r.IndexerName,
		downloadURL:   r.DownloadURL,
		title:         r.Title,
		format:        format,
		size:          r.Size,
		indexerResult: r,
	}
}

// scoreResults calculates quality scores for all results and returns them sorted by score descending.
// Scoring considers: format preference, file size, provider reliability, seeders/peers (indexers only),
// and quality profile preferences.
func scoreResults(results []*unifiedResult, profile *qualityprofile.QualityProfile) []*unifiedResult {
	for _, r := range results {
		r.score = calculateUnifiedScore(r, profile)
	}

	// Sort by score descending (highest quality first)
	// Using simple bubble sort for clarity - can optimize later if needed
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// calculateUnifiedScore assigns a quality score to a unified result.
// Higher scores indicate better quality sources.
func calculateUnifiedScore(r *unifiedResult, profile *qualityprofile.QualityProfile) int {
	score := 0
	format := strings.ToLower(r.format)

	// 1. Format preference with quality profile integration
	formatScore := scoreFormat(format, profile)
	score += formatScore

	// 2. Provider reliability bonus
	if r.isLibrary {
		score += scoreLibraryProvider(r.sourceName)
	} else if r.isIndexer {
		// Indexer sources (Prowlarr/Usenet) get a high baseline score
		// because they typically have retail/publisher quality
		score += 150 // Base score for indexer sources

		// Add seeders/peers bonus (torrent popularity = quality signal)
		if r.indexerResult != nil {
			score += scoreSeedersAndPeers(r.indexerResult)
		}
	}

	// 3. File size quality signal
	// Larger files typically indicate better quality (CSS, images, formatting preserved)
	// But avoid extreme sizes that might be malformed/corrupted
	score += scoreFileSize(r.size, format)

	// 4. Metadata completeness (library only)
	if r.isLibrary && r.libraryResult != nil {
		lr := r.libraryResult
		if len(lr.Authors) > 0 {
			score += 10
		}
		if lr.ISBN != "" {
			score += 8
		}
		if lr.Year > 0 {
			if lr.Year >= 2020 {
				score += 20
			} else if lr.Year >= 2010 {
				score += 10
			}
		}
	}

	return score
}

// scoreFormat scores a format based on preference and quality profile.
func scoreFormat(format string, profile *qualityprofile.QualityProfile) int {
	// Base format scores
	var base int
	switch format {
	case "epub":
		base = 100
	case "mobi", "azw", "azw3":
		base = 70
	case "pdf":
		base = 50
	default:
		base = 20
	}

	// Apply quality profile multipliers
	if profile != nil {
		// Check if format is allowed
		allowed := false
		for _, item := range profile.Items {
			if strings.EqualFold(item.Quality, format) && item.Allowed {
				allowed = true
				break
			}
		}
		if !allowed {
			return 0 // Disallowed format gets zero score
		}

		// Bonus if this format meets or exceeds cutoff
		if strings.EqualFold(profile.Cutoff, format) {
			base += 50 // Cutoff format gets significant bonus
		}
	}

	return base
}

// scoreLibraryProvider scores library providers based on reliability.
func scoreLibraryProvider(provider string) int {
	provider = strings.ToLower(provider)
	switch provider {
	case "gutendex":
		return 45
	case "wikisource":
		return 42
	case "aozora":
		return 40
	case "openlibrary-public":
		return 40
	case "internet-archive":
		return 40
	case "dominio-publico":
		return 35
	case "libgen":
		return 30 // LibGen often has lower-quality conversions
	case "annas-archive":
		return 25 // Anna's Archive similar quality to LibGen
	case "gutenberg-au", "gutenberg-ca", "biblioteca-digital-hispanica", "gallica",
		"projekt-gutenberg-de", "baen-free-library", "ccel", "sefaria", "ctext",
		"sacred-texts", "digital-comic-museum", "hathitrust":
		return 30 // Public catalogs via site search
	default:
		return 10
	}
}

// scoreSeedersAndPeers scores torrent health based on seeders and peers.
// More seeders = more popular = likely better quality release.
func scoreSeedersAndPeers(r *search.Result) int {
	if r.Seeders == 0 {
		return 0 // Dead torrent
	}

	score := 0
	// Seeders are the primary quality signal
	switch {
	case r.Seeders >= 50:
		score += 80 // Very popular, likely high quality
	case r.Seeders >= 20:
		score += 60
	case r.Seeders >= 10:
		score += 40
	case r.Seeders >= 5:
		score += 25
	default:
		score += 10
	}

	// Leechers indicate current activity
	if r.Leechers > 0 {
		score += 5
	}

	// Grabs indicate historical popularity
	if r.Grabs >= 100 {
		score += 15
	} else if r.Grabs >= 50 {
		score += 10
	} else if r.Grabs >= 10 {
		score += 5
	}

	return score
}

// scoreFileSize scores based on file size as a quality signal.
// Larger files typically have better formatting, but extremely large files might be corrupted.
func scoreFileSize(size int64, format string) int {
	if size == 0 {
		return 0
	}

	sizeMB := float64(size) / (1024 * 1024)

	// Different formats have different expected size ranges
	switch format {
	case "epub", "mobi", "azw", "azw3":
		// EPUB/MOBI: typical retail book is 2-10 MB for text-heavy, 10-50 MB with images
		switch {
		case sizeMB >= 10 && sizeMB <= 100: // Rich content (images, CSS, fonts)
			return 40
		case sizeMB >= 5 && sizeMB < 10: // Good content
			return 35
		case sizeMB >= 2 && sizeMB < 5: // Decent content
			return 25
		case sizeMB >= 1 && sizeMB < 2: // Minimal content
			return 15
		case sizeMB < 1: // Very small, likely low quality
			return 5
		default: // Over 100 MB, unusual but not necessarily bad
			return 20
		}
	case "pdf":
		// PDF: can be much larger due to scanned images
		switch {
		case sizeMB >= 20 && sizeMB <= 200: // Rich PDF
			return 35
		case sizeMB >= 10 && sizeMB < 20: // Good PDF
			return 30
		case sizeMB >= 5 && sizeMB < 10: // Decent PDF
			return 20
		case sizeMB < 5: // Small PDF, likely low quality scan
			return 10
		default: // Very large
			return 15
		}
	default:
		// Generic scoring for other formats
		if sizeMB >= 1 {
			return 20
		}
		return 10
	}
}

// extractFormat attempts to extract the format from a title or filename.
func extractFormat(title string) string {
	lower := strings.ToLower(title)
	formats := []string{"epub", "mobi", "azw3", "azw", "pdf", "cbz", "cbr"}
	for _, fmt := range formats {
		if strings.Contains(lower, "."+fmt) || strings.Contains(lower, " "+fmt+" ") || strings.HasSuffix(lower, " "+fmt) {
			return fmt
		}
	}
	return "unknown"
}

// filterByQualityProfile filters results to only allowed formats.
func filterByQualityProfile(results []*unifiedResult, profile *qualityprofile.QualityProfile) []*unifiedResult {
	if profile == nil {
		return results // No filtering
	}

	var filtered []*unifiedResult
	for _, r := range results {
		format := strings.ToLower(r.format)
		allowed := false
		for _, item := range profile.Items {
			if strings.EqualFold(item.Quality, format) && item.Allowed {
				allowed = true
				break
			}
		}
		if allowed {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// String returns a human-readable representation of the result for logging.
func (r *unifiedResult) String() string {
	sizeMB := float64(r.size) / (1024 * 1024)
	source := "library"
	if r.isIndexer {
		source = "indexer"
	}
	return fmt.Sprintf("[%s:%s] %s %.1fMB (score: %d)", source, r.sourceName, r.format, sizeMB, r.score)
}
