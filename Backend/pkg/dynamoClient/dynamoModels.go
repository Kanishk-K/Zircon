package dynamo

type UserDocument struct {
	UserID               string   `dynamodbav:"userID"`
	CreatedOn            string   `dynamodbav:"createdOn"`
	PermittedGenerations int      `dynamodbav:"permittedGenerations"`
	ScheduledJobs        []string `dynamodbav:"scheduledJobs,stringset,omitempty"`
	Name                 string   `dynamodbav:"name"`
}

type JobDocument struct {
	EntryID            string   `dynamodbav:"entryID"`
	GeneratedOn        string   `dynamodbav:"generatedOn"`
	GeneratedBy        string   `dynamodbav:"generatedBy"`
	SubtitlesGenerated bool     `dynamodbav:"subtitlesGenerated"`
	VideosAvailable    []string `dynamodbav:"videosAvailable,stringset,omitempty"`
}

type VideoRequestDocument struct {
	RequestKey     string `dynamodbav:"requestKey"`
	EntryID        string `dynamodbav:"entryID"`
	RequestedVideo string `dynamodbav:"requestedVideo"`
	RequestedOn    string `dynamodbav:"requestedOn"`
	RequestedBy    string `dynamodbav:"requestedBy"`
	VideoExpiry    int    `dynamodbav:"videoExpiry"`
}
