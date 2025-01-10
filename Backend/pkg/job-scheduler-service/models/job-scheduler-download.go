package models

import "github.com/hibiken/asynq"

type JobQueueRequest struct {
	EntryID         string `json:"entryID"`
	UserID          string `json:"userID"`
	TranscriptLink  string `json:"transcript"`
	Title           string `json:"title"`
	Notes           bool   `json:"notes"`
	Summarize       bool   `json:"summarize"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type JobStatusRequest struct {
	EntryID string `json:"entryID"`
}

type JobStatusResponse struct {
	SummarizeStatus asynq.TaskState `json:"summarizeStatus"`
	VideoStatus     asynq.TaskState `json:"videoStatus"`
}

type SummarizeInformation struct {
	EntryID         string `json:"entryID"`
	TranscriptLink  string `json:"transcript"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type GenerateVideoInformation struct {
	EntryID           string `json:"entryID"`
	BackgroundVideo   string `json:"backgroundVideo"`
	GenerateSubtitles bool   `json:"generateSubtitles"`
}
