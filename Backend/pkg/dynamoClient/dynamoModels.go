package dynamo

type UserDocument struct {
	UserID               string   `dynamodbav:"userID"`
	CreatedOn            string   `dynamodbav:"createdOn"`
	PermittedGenerations int      `dynamodbav:"permittedGenerations"`
	ScheduledJobs        []string `dynamodbav:"scheduledJobs,stringset,omitempty"`
}

type JobDocument struct {
	EntryID            string   `dynamodbav:"entryID"`
	Title              string   `dynamodbav:"title"`
	GeneratedOn        string   `dynamodbav:"generatedOn"`
	GeneratedBy        string   `dynamodbav:"generatedBy"`
	SubtitlesGenerated bool     `dynamodbav:"subtitlesGenerated"`
	VideosAvailable    []string `dynamodbav:"videosAvailable,stringset,omitempty"`
}

type VideoRequestDocument struct {
	EntryID        string `dynamodbav:"entryID"`
	RequestedVideo string `dynamodbav:"requestedVideo"`
	RequestedOn    string `dynamodbav:"requestedOn"`
	RequestedBy    string `dynamodbav:"requestedBy"`
	VideoExpiry    int    `dynamodbav:"videoExpiry"`
}
