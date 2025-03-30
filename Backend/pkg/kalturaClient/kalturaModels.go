package kalturaclient

// KalturaSessionResponse represents the response for starting a widget session.
// This is always the first returned object in the Kaltura API response.
type KalturaSessionResponse struct {
	PartnerID  int    `json:"partnerId"`
	Ks         string `json:"ks"`
	UserID     int    `json:"userId"`
	ObjectType string `json:"objectType"`
}

// KalturaTranscriptAsset represents a transcript asset in Kaltura.
type KalturaTranscriptAsset struct {
	ID            string `json:"id"`
	EntryID       string `json:"entryId"`
	PartnerID     int    `json:"partnerId"`
	Version       string `json:"version"`
	Size          int    `json:"size"`
	Tags          string `json:"tags"`
	FileExt       string `json:"fileExt"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
	Description   string `json:"description"`
	SizeInBytes   string `json:"sizeInBytes"`
	Filename      string `json:"filename"`
	Format        string `json:"format"`
	Status        int    `json:"status"`
	Accuracy      int    `json:"accuracy"`
	HumanVerified bool   `json:"humanVerified"`
	Language      string `json:"language"`
	ObjectType    string `json:"objectType"`
}

// KalturaAttachmentAssetListResponse represents a list response for attachment assets.
// This is the second returned object in the Kaltura API response.
type KalturaAttachmentAssetListResponse struct {
	TotalCount int                      `json:"totalCount"`
	Objects    []KalturaTranscriptAsset `json:"objects"`
	ObjectType string                   `json:"objectType"`
}