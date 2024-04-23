package helpers

// UserURL представляет структуру данных для хранения информации о URL-адресах пользователя.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
