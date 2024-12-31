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
	Summary         string `json:"summary"` // TODO: Change to S3 URL
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}
