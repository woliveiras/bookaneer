package openlibrary

type authorSearchResponse struct {
	NumFound int         `json:"numFound"`
	Start    int         `json:"start"`
	Docs     []authorDoc `json:"docs"`
}

type authorDoc struct {
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	AlternateNames []string `json:"alternate_names"`
	BirthDate      int      `json:"birth_date"`
	DeathDate      int      `json:"death_date"`
	TopWork        string   `json:"top_work"`
	WorkCount      int      `json:"work_count"`
	TopSubjects    []string `json:"top_subjects"`
}

type bookSearchResponse struct {
	NumFound int       `json:"numFound"`
	Start    int       `json:"start"`
	Docs     []bookDoc `json:"docs"`
}

type bookDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	AuthorKey        []string `json:"author_key"`
	FirstPublishYear int      `json:"first_publish_year"`
	CoverI           int      `json:"cover_i"`
	ISBN             []string `json:"isbn"`
	Language         []string `json:"language"`
	Subject          []string `json:"subject"`
	Publisher        []string `json:"publisher"`
	EditionCount     int      `json:"edition_count"`
}

type authorResponse struct {
	Key          string      `json:"key"`
	Name         string      `json:"name"`
	PersonalName string      `json:"personal_name"`
	Bio          interface{} `json:"bio"`
	BirthDate    string      `json:"birth_date"`
	DeathDate    string      `json:"death_date"`
	Wikipedia    string      `json:"wikipedia"`
	Photos       []int       `json:"photos"`
	Links        []linkRef   `json:"links"`
}

type linkRef struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  struct {
		Key string `json:"key"`
	} `json:"type"`
}

type workResponse struct {
	Key              string      `json:"key"`
	Title            string      `json:"title"`
	Subtitle         string      `json:"subtitle"`
	Description      interface{} `json:"description"`
	Covers           []int       `json:"covers"`
	Subjects         []string    `json:"subjects"`
	SubjectPlaces    []string    `json:"subject_places"`
	SubjectTimes     []string    `json:"subject_times"`
	SubjectPeople    []string    `json:"subject_people"`
	Authors          []authorRef `json:"authors"`
	FirstPublishDate string      `json:"first_publish_date"`
	Links            []linkRef   `json:"links"`
}

type authorRef struct {
	Type struct {
		Key string `json:"key"`
	} `json:"type"`
	Author struct {
		Key string `json:"key"`
	} `json:"author"`
}

type editionResponse struct {
	Key            string      `json:"key"`
	Title          string      `json:"title"`
	Subtitle       string      `json:"subtitle"`
	Publishers     []string    `json:"publishers"`
	PublishDate    string      `json:"publish_date"`
	NumberOfPages  int         `json:"number_of_pages"`
	Covers         []int       `json:"covers"`
	ISBN10         []string    `json:"isbn_10"`
	ISBN13         []string    `json:"isbn_13"`
	LCCN           []string    `json:"lccn"`
	OCLC           []string    `json:"oclc_numbers"`
	Languages      []keyRef    `json:"languages"`
	Works          []keyRef    `json:"works"`
	Authors        []keyRef    `json:"authors"`
	Description    interface{} `json:"description"`
	PhysicalFormat string      `json:"physical_format"`
}

type keyRef struct {
	Key string `json:"key"`
}
