package dominiopublico

import "testing"

func TestParseSearchResults(t *testing.T) {
	provider := New()
	html := `
	<html><body>
		<a href="/pesquisa/DetalheObraForm.do?select_action=&co_obra=12345">Dom Casmurro</a>
		<a href="/pesquisa/DetalheObraForm.do?select_action=&co_obra=67890"><span>Memórias Póstumas</span></a>
	</body></html>`

	results := provider.parseSearchResults(html)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].ID != "12345" {
		t.Fatalf("expected id 12345, got %s", results[0].ID)
	}
	if results[0].Title != "Dom Casmurro" {
		t.Fatalf("unexpected title: %s", results[0].Title)
	}
	if results[0].Provider != "dominio-publico" {
		t.Fatalf("unexpected provider: %s", results[0].Provider)
	}
}

func TestExtractBestFileURL(t *testing.T) {
	html := `
	<html><body>
		<a href="/download/livro.pdf">PDF</a>
		<a href="/download/livro.epub">EPUB</a>
	</body></html>`

	best := extractBestFileURL(html)
	if best != "/download/livro.epub" {
		t.Fatalf("expected epub url, got %s", best)
	}
}
