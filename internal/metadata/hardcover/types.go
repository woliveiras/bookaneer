package hardcover

type searchResponse struct {
	Data struct {
		Search struct {
			Results []searchResult `json:"results"`
		} `json:"search"`
	} `json:"data"`
}

type searchResult struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	AuthorNames []string `json:"author_names"`
	ReleaseYear int      `json:"release_year"`
	ISBNs       []string `json:"isbns"`
	BooksCount  int      `json:"books_count"`
	Slug        string   `json:"slug"`
	Image       imageObj `json:"image"`
}

type imageObj struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Color  string `json:"color"`
}

type authorResponse struct {
	Data struct {
		Authors []authorObj `json:"authors"`
	} `json:"data"`
}

type authorObj struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Bio   string   `json:"bio"`
	Slug  string   `json:"slug"`
	Image imageObj `json:"image"`
}

type bookResponse struct {
	Data struct {
		Books []bookObj `json:"books"`
	} `json:"data"`
}

type bookObj struct {
	ID            int               `json:"id"`
	Title         string            `json:"title"`
	Subtitle      string            `json:"subtitle"`
	Description   string            `json:"description"`
	ReleaseDate   string            `json:"release_date"`
	Pages         int               `json:"pages"`
	Slug          string            `json:"slug"`
	CachedImage   imageObj          `json:"cached_image"`
	Contributions []contributionObj `json:"contributions"`
	BookSeries    []bookSeriesObj   `json:"book_series"`
	Taggings      []taggingObj      `json:"taggings"`
}

type contributionObj struct {
	Author struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"author"`
}

type bookSeriesObj struct {
	Series struct {
		Name string `json:"name"`
	} `json:"series"`
	Position float64 `json:"position"`
}

type taggingObj struct {
	Tag struct {
		Tag string `json:"tag"`
	} `json:"tag"`
}

type editionResponse struct {
	Data struct {
		Editions []editionObj `json:"editions"`
	} `json:"data"`
}

type editionObj struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	ISBN10 string `json:"isbn_10"`
	ISBN13 string `json:"isbn_13"`
	Book   struct {
		ID int `json:"id"`
	} `json:"book"`
}
