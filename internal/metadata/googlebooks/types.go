package googlebooks

type volumesResponse struct {
	Kind       string       `json:"kind"`
	TotalItems int          `json:"totalItems"`
	Items      []volumeItem `json:"items"`
}

type volumeItem struct {
	Kind       string     `json:"kind"`
	ID         string     `json:"id"`
	Etag       string     `json:"etag"`
	SelfLink   string     `json:"selfLink"`
	VolumeInfo volumeInfo `json:"volumeInfo"`
	SaleInfo   saleInfo   `json:"saleInfo"`
	AccessInfo accessInfo `json:"accessInfo"`
}

type volumeInfo struct {
	Title               string       `json:"title"`
	Subtitle            string       `json:"subtitle"`
	Authors             []string     `json:"authors"`
	Publisher           string       `json:"publisher"`
	PublishedDate       string       `json:"publishedDate"`
	Description         string       `json:"description"`
	IndustryIdentifiers []industryID `json:"industryIdentifiers"`
	PageCount           int          `json:"pageCount"`
	PrintType           string       `json:"printType"`
	Categories          []string     `json:"categories"`
	AverageRating       float64      `json:"averageRating"`
	RatingsCount        int          `json:"ratingsCount"`
	MaturityRating      string       `json:"maturityRating"`
	ContentVersion      string       `json:"contentVersion"`
	ImageLinks          imageLinks   `json:"imageLinks"`
	Language            string       `json:"language"`
	PreviewLink         string       `json:"previewLink"`
	InfoLink            string       `json:"infoLink"`
	CanonicalVolumeLink string       `json:"canonicalVolumeLink"`
	SeriesInfo          *seriesInfo  `json:"seriesInfo"`
}

type industryID struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type imageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
	Small          string `json:"small"`
	Medium         string `json:"medium"`
	Large          string `json:"large"`
	ExtraLarge     string `json:"extraLarge"`
}

type seriesInfo struct {
	Kind              string `json:"kind"`
	BookDisplayNumber string `json:"bookDisplayNumber"`
	VolumeSeries      []struct {
		SeriesID       string `json:"seriesId"`
		SeriesBookType string `json:"seriesBookType"`
		OrderNumber    int    `json:"orderNumber"`
	} `json:"volumeSeries"`
}

type saleInfo struct {
	Country     string `json:"country"`
	Saleability string `json:"saleability"`
	IsEbook     bool   `json:"isEbook"`
	ListPrice   *price `json:"listPrice"`
	RetailPrice *price `json:"retailPrice"`
	BuyLink     string `json:"buyLink"`
}

type price struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}

type accessInfo struct {
	Country                string       `json:"country"`
	Viewability            string       `json:"viewability"`
	Embeddable             bool         `json:"embeddable"`
	PublicDomain           bool         `json:"publicDomain"`
	TextToSpeechPermission string       `json:"textToSpeechPermission"`
	Epub                   formatAccess `json:"epub"`
	Pdf                    formatAccess `json:"pdf"`
	WebReaderLink          string       `json:"webReaderLink"`
	AccessViewStatus       string       `json:"accessViewStatus"`
}

type formatAccess struct {
	IsAvailable  bool   `json:"isAvailable"`
	DownloadLink string `json:"downloadLink"`
	AcsTokenLink string `json:"acsTokenLink"`
}
