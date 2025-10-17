package models

// Note представляет заметку пользователя
type Note struct {
	ID        uint64 `json:"id"`
	OwnerID   uint64 `json:"owner_id"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Favourite bool   `json:"favorite"`
	Folder    string `json:"folder"`
}
