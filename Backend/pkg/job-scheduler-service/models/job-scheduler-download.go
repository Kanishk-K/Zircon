package models

import "github.com/hibiken/asynq"

type JobQueueRequest struct {
	EntryID         string `json:"entryID"`
	UserID          string `json:"userID"`
	TranscriptLink  string `json:"transcript"`
	Notes           bool   `json:"notes"`
	Summarize       bool   `json:"summarize"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type JobStatusRequest struct {
	EntryID string `json:"entryID"`
}

type JobStatusResponse struct {
	NotesStatus     asynq.TaskState `json:"notesStatus"`
	SummarizeStatus asynq.TaskState `json:"summarizeStatus"`
	VideoStatus     asynq.TaskState `json:"videoStatus"`
}

type JobExistingResponse struct {
	NotesGenerated   bool     `json:"notesGenerated"`
	SummaryGenerated bool     `json:"summaryGenerated"`
	VideosAvailable  []string `json:"videosAvailable"`
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

type NotesInformation struct {
	EntryID        string `json:"entryID"`
	TranscriptLink string `json:"transcript"`
}
