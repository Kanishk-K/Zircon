package models

type UserDocument struct {
	UserID               string   `dynamodbav:"userID"`
	CreatedOn            string   `dynamodbav:"createdOn"`
	PermittedGenerations int      `dynamodbav:"permittedGenerations"`
	ScheduledJobs        []string `dynamodbav:"scheduledJobs"`
	Name                 string   `dynamodbav:"name"`
}

type JobDocument struct {
	EntryID            string   `dynamodbav:"entryID"`
	GeneratedOn        string   `dynamodbav:"generatedOn"`
	GeneratedBy        string   `dynamodbav:"generatedBy"`
	NotesGenerated     bool     `dynamodbav:"notesGenerated"`
	SummaryGenerated   bool     `dynamodbav:"summaryGenerated"`
	SubtitlesGenerated bool     `dynamodbav:"subtitlesGenerated"`
	VideosAvailable    []string `dynamodbav:"videosAvailable,stringset,omitempty"`
}
