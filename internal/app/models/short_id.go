package models

// ShortURL is main entity for system.
type ShortURL struct {
	OriginalURL string `json:"url"`        // original URL that was shortened
	ShortURL    string `json:"id"`         // unique ShortURL of the short URL.
	CreatedByID string `json:"created_by"` // ShortURL of the user who created the short URL
	IsDeleted   bool   `json:"is_deleted"` // is used to mark a record as deleted
}
