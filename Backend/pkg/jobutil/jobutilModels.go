package jobutil

type JobQueueRequest struct {
	EntryID         string `json:"entryID"`
	Title           string `json:"title"`
	TranscriptLink  string `json:"transcript"`
	BackgroundVideo string `json:"backgroundVideo"`
}
