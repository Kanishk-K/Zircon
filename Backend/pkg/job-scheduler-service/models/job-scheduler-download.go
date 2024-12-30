package models

type JobInformation struct {
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	Notes           bool   `json:"notes"`
	Summarize       bool   `json:"summarize"`
	BackgroundVideo string `json:"backgroundVideo"`
}
