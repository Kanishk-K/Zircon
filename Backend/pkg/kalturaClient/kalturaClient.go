package kalturaclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func BuildQueryURL(entryID string) string {
	// Build the URL for the Kaltura API
	baseURL := "https://cdnapi.kaltura.com/api_v3/index.php"
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Add("service", "multirequest")
	q.Add("format", "1")
	q.Add("ignoreNull", "1")
	q.Add("1:service", "session")
	q.Add("1:action", "startWidgetSession")
	q.Add("1:widgetId", fmt.Sprintf("_%s", os.Getenv("KALTURA_PARTNER_ID")))
	q.Add("2:ks", "{1:result:ks}")
	q.Add("2:service", "attachment_attachmentAsset")
	q.Add("2:action", "list")
	q.Add("2:filter:entryIdEqual", entryID)

	u.RawQuery = q.Encode()
	return u.String()
}

func BuildTranscriptLinkURL(assetID string) string {
	// Build the URL for the Kaltura API
	return fmt.Sprintf("https://cdnapi.kaltura.com/api_v3/index.php/service/attachment_attachmentAsset/action/serve/attachmentAssetId/%s", assetID)
}

func GetTranscriptLink(entryID string) (string, error) {
	// Need to make an API call to the Kaltura API to get the transcript link
	resp, err := http.Get(BuildQueryURL(entryID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	// Parse the response to get the transcript link
	var rawResponses []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawResponses); err != nil {
		return "", err
	}
	var kalturaAttachmentResponse KalturaAttachmentAssetListResponse
	if err := json.Unmarshal(rawResponses[1], &kalturaAttachmentResponse); err != nil {
		return "", fmt.Errorf("failed to parse Kaltura Attachment response")
	}
	if kalturaAttachmentResponse.TotalCount == 0 {
		return "", fmt.Errorf("no transcript found for entry ID %s", entryID)
	}
	for _, asset := range kalturaAttachmentResponse.Objects {
		if strings.HasSuffix(asset.Filename, ".txt") && strings.Contains(asset.Filename, entryID) {
			return BuildTranscriptLinkURL(asset.ID), nil
		}
	}
	return "", fmt.Errorf("no transcript found for entry ID %s", entryID)
}
