package jobutil

type JobQueueRequest struct {
	EntryID         string `json:"entryID"`
	Title           string `json:"title"`
	BackgroundVideo string `json:"backgroundVideo"`
}
