package tasks

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/hibiken/asynq"
)

const TypeTTSSummary = "summary:tts"

type TTSSummaryProcess struct {
	PollyClient *polly.Polly
}

func NewTTSSummaryProcess(client *polly.Polly) *TTSSummaryProcess {
	return &TTSSummaryProcess{
		PollyClient: client,
	}
}

func NewTTSSummaryTask(jobInfo *models.TTSSummaryInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTTSSummary, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
}

type subtitleWord struct {
	Time  int    `json:"time"`
	Value string `json:"value"`
}

type subtitleLine struct {
	Start int
	End   int
	Text  string
}

const CHARSPERLINE = 25
const TEMPOSPEED = 1.25 // 1.25x speed should match atempo=1.25 in ffmpeg

func (p *TTSSummaryProcess) HandleTTSSummaryTask(ctx context.Context, t *asynq.Task) error {
	data := models.TTSSummaryInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		return err
	}
	log.Printf("Tasked to generate TTS summary for: %s", data.Title)
	// This generates the TTS Audio
	/*
		TTSMp3Input := &polly.SynthesizeSpeechInput{
			OutputFormat: aws.String("mp3"),
			Text:         aws.String(data.Summary),
			VoiceId:      aws.String("Joey"),
			Engine:       aws.String("standard"),
		}


			Mp3Output, err := p.PollyClient.SynthesizeSpeech(TTSMp3Input)
			if err != nil {
				log.Printf("API to AWS Polly (Speech) Failed: %v", err)
				return err
			}

			Mp3Fp, err := os.CreateTemp("", "tts-*.mp3")
			if err != nil {
				log.Printf("Failed to create temp file: %v", err)
				return err
			}
			defer Mp3Fp.Close()

			_, err = io.Copy(Mp3Fp, Mp3Output.AudioStream)
			if err != nil {
				log.Printf("Failed to write to temp file: %v", err)
				return err
			}
	*/

	// This generates the TTS Subtitles
	TTSSubtitlesInput := &polly.SynthesizeSpeechInput{
		OutputFormat:    aws.String("json"),
		Text:            aws.String(data.Summary),
		VoiceId:         aws.String("Joey"),
		Engine:          aws.String("standard"),
		SpeechMarkTypes: aws.StringSlice([]string{"word"}),
	}

	SubtitlesOutput, err := p.PollyClient.SynthesizeSpeech(TTSSubtitlesInput)
	if err != nil {
		log.Printf("API to AWS Polly (Subtitle) Failed: %v", err)
		return err
	}

	// Convert the TTS Subtitles to ASS Format
	AssFp, err := os.CreateTemp("", "subtitle-*.ass")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return err
	}
	// defer os.Remove(AssFp.Name())
	defer AssFp.Close()

	words, err := ParseJSON(SubtitlesOutput.AudioStream)
	if err != nil {
		log.Printf("Failed to parse subtitles")
		return nil
	}

	Mp3Fp, err := os.Open("c:/Users/Kanis/Desktop/tts.mp3")
	if err != nil {
		log.Printf("Unable to open Subtitle File")
		return err
	}
	defer Mp3Fp.Close()

	maxDur, err := GetMaxMp3Duration(Mp3Fp)
	if err != nil {
		log.Printf("Failed to get Mp3 Duration")
		return err
	}

	subtitleLines := GenerateSubtitleLines(words, maxDur)

	if err := GenerateAAS(subtitleLines, AssFp); err != nil {
		log.Printf("Failed to insert subtitles")
		return err
	}

	log.Printf("Generated TTS summary for: %s", data.Title)

	return nil
}

func ParseJSON(subtitleReader io.ReadCloser) ([]subtitleWord, error) {

	scanner := bufio.NewScanner(subtitleReader)
	var subtitles []subtitleWord

	for scanner.Scan() {
		line := scanner.Text()
		var word subtitleWord
		if err := json.Unmarshal([]byte(line), &word); err != nil {
			log.Printf("Error parsing the line: %s\n", err)
			return nil, err
		}
		subtitles = append(subtitles, word)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %s\n", err)
		return nil, err
	}

	return subtitles, nil
}

func GetMaxMp3Duration(Mp3Ptr *os.File) (int, error) {
	// Call ffprobe to get the duration of the mp3 file
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 input.mp3
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", Mp3Ptr.Name())
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get the duration of the mp3 file: %v", err)
		return 0, err
	}
	// Convert the output (float) to int rounded down
	durationStr := strings.TrimSpace(string(out))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		log.Printf("Failed to convert the duration to int: %v", err)
		return 0, err
	}
	// Return the duration in milliseconds
	return int(duration * 1000), nil
}

func GenerateSubtitleLines(words []subtitleWord, maxTime int) []subtitleLine {
	// Create a function that will take in the words and generate subtitle lines that do not exceed CHARSPERLINE
	// When adding the first word to a line, set the start time of the line as the time of the word
	// When a line cannot add any more words set its end as the time of the upcoming word
	// If there are no more words, set the end as the maxTime
	var lines []subtitleLine
	var line subtitleLine
	var lineText string
	for i, word := range words {
		if i == 0 {
			line.Start = word.Time
			lineText = word.Value
		} else {
			if len(lineText)+len(word.Value)+1 <= CHARSPERLINE {
				lineText += " " + word.Value
			} else {
				line.End = word.Time - 5
				line.Text = lineText
				lines = append(lines, line)
				line = subtitleLine{}
				line.Start = word.Time + 5
				lineText = word.Value
			}
		}
		if i == len(words)-1 {
			line.End = maxTime
			line.Text = lineText
			lines = append(lines, line)
		}
	}
	return lines
}

func (line subtitleLine) timeAsString(time int) string {
	floatTime := float64(time) / TEMPOSPEED
	time = int(floatTime)

	hours := time / 3600000
	time = time % 3600000

	minutes := time / 60000
	time = time % 60000

	seconds := time / 1000
	time = time % 1000

	hundredths := time / 10

	return fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, hundredths)

}

func (line subtitleLine) startAsString() string {
	return line.timeAsString(line.Start)
}

func (line subtitleLine) endAsString() string {
	return line.timeAsString(line.End)
}

func GenerateAAS(lines []subtitleLine, aasFptr *os.File) error {
	// Write the Static Script Info Header
	aasFptr.WriteString("[Script Info]\n")
	aasFptr.WriteString("PlayResX: 1080\n")
	aasFptr.WriteString("PlayResY: 1920\n")
	aasFptr.WriteString("WrapStyle: 0\n\n")

	// Write the V4+ Styles Header
	aasFptr.WriteString("[V4+ Styles]\n")
	aasFptr.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	aasFptr.WriteString("Style: Default,Impact,50,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,4,4,2,10,10,10,1\n\n")

	// Write The Events Header
	aasFptr.WriteString("[Events]\n")
	aasFptr.WriteString("Format: Layer, Start, End, Style, Text\n")
	for _, line := range lines {
		aasFptr.WriteString("Dialogue: 0," + line.startAsString() + "," + line.endAsString() + ",Default,{\\an5\\pos(540,960)\\fscx160\\fscy160\\alpha&HFF&\\t(0,75,\\alpha&H00&)\\t(0,75,\\fscx220\\fscy220)\\t(75,150,\\fscx200\\fscy200)\\}" + line.Text + "\n")
	}
	return nil
}
