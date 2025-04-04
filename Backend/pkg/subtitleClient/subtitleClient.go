package subtitleclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type SubtitleGenerationMethods interface {
	GenerateTTS(textInput string) (*LemonFoxResponse, error)
}

type SubtitleClient struct {
	baseURL string
}

func NewSubtitleClient() SubtitleGenerationMethods {
	return &SubtitleClient{
		baseURL: "https://api.lemonfox.ai/v1",
	}
}

func (sc *SubtitleClient) GenerateTTS(textInput string) (*LemonFoxResponse, error) {
	body, err := json.Marshal(
		map[string]interface{}{
			"input": textInput,
			"voice": "adam",
			// "speed":           1.2,
			"word_timestamps": true,
			"response_format": "aac",
		},
	)
	if err != nil {
		log.Println("Failed to marshal request body")
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, sc.baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		log.Println("Failed to create request")
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("LEMONFOX_API_KEY"))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to make request")
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Erroneous response code: ", resp.StatusCode)
		return nil, fmt.Errorf("erroneous response code: %d", resp.StatusCode)
	}
	LemonFoxResponse := &LemonFoxResponse{}
	err = json.NewDecoder(resp.Body).Decode(LemonFoxResponse)
	if err != nil {
		log.Println("Failed to decode response body")
		return nil, err
	}
	return LemonFoxResponse, nil
}

func isPunctuation(word WordTimeStamp) bool {
	return word.Word == "." || word.Word == "," || word.Word == "?" || word.Word == "!" || word.Word == ":" || word.Word == ";"
}

func ConvertB64ToAudio(b64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		log.Println("Failed to decode base64 string")
		return nil, err
	}
	return data, nil
}

func GenerateSubtitleLines(words []WordTimeStamp) []LineTimeStamp {
	var lines []LineTimeStamp
	line := []WordTimeStamp{}
	strlen := 0
	for i, word := range words {
		if i == 0 {
			// First word, initialize line
			line = append(line, word)
			strlen = len(word.Word)
			continue
		}
		if strlen+len(word.Word)+1 <= MAX_CHARS_PER_LINE {
			// Add word to line
			line = append(line, word)
			strlen += len(word.Word) + 1 // Add 1 for prepended space
		} else {
			// Line is full, but this word is just punctuation so add it to this line
			if isPunctuation(word) {
				line = append(line, word)
				strlen += len(word.Word)
			} else {
				// Add line to lines
				lines = append(lines, LineTimeStamp{
					Line: line,
				})
				// Start new line
				line = []WordTimeStamp{word}
				strlen = len(word.Word)
			}
		}
	}
	// Clean up last line
	lines = append(lines, LineTimeStamp{
		Line: line,
	})
	return lines
}

func GenerateASSContent(lines []LineTimeStamp) string {
	/* STATIC HEADER
	* [Script Info]
	* PlayResX: 576
	* PlayResY: 1024
	* WrapStyle: 0
	*
	* [V4+ Styles]
	* Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	* Style: Default,Berlin Sans FB,50,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,4,4,2,10,10,10,1
	*
	* [Events]
	* Format: Layer, Start, End, Style, Text
	 */
	AASContent := "[Script Info]\nPlayResX: 576\nPlayResY: 1024\nWrapStyle: 0\n\n[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\nStyle: Default,Berlin Sans FB,50,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,4,4,2,10,10,10,1\n\n[Events]\nFormat: Layer, Start, End, Style, Text\n"
	for _, line := range lines {
		lineStart, lineEnd := line.Duration()
		AASContent += fmt.Sprintf("Dialogue: 0,%s,%s,Default,{\\an5\\pos(288,512)\\fscx60\\fscy60\\alpha&HFF&\\t(0,35,\\alpha&H00&)\\t(0,35,\\fscx90\\fscy90)\\t(35,75,\\fscx70\\fscy70)}", lineStart, lineEnd)
		for i, word := range line.Line {
			// Format should generate as follows: {\1c&HFFFFFF&\t(start,start,HIGHLIGHT_COLOR)\t(end,end,\1c&HFFFFFF&)}Word
			// start is the time since the beginning of the line (in ms)
			// end is the time since the beginning of the line (in ms)
			startOffset := int((word.StartTime - line.Line[0].StartTime) * 1000)
			endOffset := int((word.EndTime - line.Line[0].StartTime) * 1000)
			if i == 0 {
				// First word in line, make it start with the highlight color
				AASContent += fmt.Sprintf("{%s\\t(%d,%d,\\1c&HFFFFFF&)}%s", HIGHLIGHT_COLOR, endOffset, endOffset, word.Word)
				continue
			}
			if !isPunctuation(word) {
				// Not punctuation, so add a space
				AASContent += " "
			}
			AASContent += fmt.Sprintf("{\\1c&HFFFFFF&\\t(%d,%d,%s)\\t(%d,%d,\\1c&HFFFFFF&)}%s", startOffset, startOffset, HIGHLIGHT_COLOR, endOffset, endOffset, word.Word)
		}
		AASContent += "\n"
	}

	return AASContent
}
