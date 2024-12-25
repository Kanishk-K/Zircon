package models

type VideoDownload struct {
	UserID    int    `json:"UserID"`
	VideoID   string `json:"VideoID"`
	SourceURL string `json:"SourceURL"`
}
