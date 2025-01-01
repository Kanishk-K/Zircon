package models

type JobInformation struct {
	EntryID         string `json:"entryID"`
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	Notes           bool   `json:"notes"`
	Summarize       bool   `json:"summarize"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type SummarizeInformation struct {
	EntryID         string `json:"entryID"`
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type TTSSummaryInformation struct {
	EntryID         string `json:"entryID"`
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}
