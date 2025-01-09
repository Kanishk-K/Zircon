package models

type JobInformation struct {
	EntryID         string `json:"entryID"`
	UserID          string `json:"userID"`
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	Notes           bool   `json:"notes"`
	Summarize       bool   `json:"summarize"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type SummarizeInformation struct {
	EntryID         string `json:"entryID"`
	UserID          string `json:"userID"`
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type GenerateVideoInformation struct {
	EntryID         string `json:"entryID"`
	UserID          string `json:"userID"`
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}
