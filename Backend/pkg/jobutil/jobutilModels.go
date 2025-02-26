package jobutil

type JobQueueRequest struct {
	EntryID         string `json:"entryID"`
	TranscriptLink  string `json:"transcript"`
	BackgroundVideo string `json:"backgroundVideo"`
}
