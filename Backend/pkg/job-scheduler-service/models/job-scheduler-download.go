package models

type JobInformation struct {
	DownloadLink  string `json:"download"`
	Title         string `json:"title"`
	Notes         bool   `json:"notes"`
	Summarize     bool   `json:"summarize"`
	Brainrot      bool   `json:"brainrot"`
	BrainrotVideo string `json:"video"`
}
