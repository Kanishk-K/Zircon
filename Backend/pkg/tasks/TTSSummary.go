package tasks

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
)

const TypeTTSSummary = "summary:tts"

type TTSSummaryProcess struct {
	PollyClient *polly.Polly
	s3Client    *s3.S3
	asynqClient *asynq.Client
}

func NewTTSSummaryProcess(client *polly.Polly, s3Client *s3.S3, asynqClient *asynq.Client) *TTSSummaryProcess {
	return &TTSSummaryProcess{
		PollyClient: client,
		s3Client:    s3Client,
		asynqClient: asynqClient,
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
	Text  []subtitleWord
}

const CHARSPERLINE = 27
const TEMPOSPEED = 1.25 // 1.25x speed should match atempo=1.25 in ffmpeg

func (p *TTSSummaryProcess) HandleTTSSummaryTask(ctx context.Context, t *asynq.Task) error {
	// Get information for the task
	data := models.TTSSummaryInformation{}
	var summary []byte

	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		return err
	}

	// [PRECHECK] : VALIDATE THAT THE SUMMARY EXISTS.
	_, err := p.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Summary.txt", data.EntryID)),
	})
	if err != nil {
		log.Printf("Failure to check if object exists, or object does not exist: %v", err)
		return err
	} else {
		// Get summary from S3
		result, err := p.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String("lecture-processor"),
			Key:    aws.String(fmt.Sprintf("%s/Summary.txt", data.EntryID)),
		})
		if err != nil {
			log.Printf("Failed to get summary from S3: %v", err)
			return err
		}
		summary, err = io.ReadAll(result.Body)
		if err != nil {
			log.Printf("Failed to read summary from S3: %v", err)
			return err
		}
	}

	// [PRECHECK] : DO NOT REPROCESS IF TTS AUDIO ALREADY EXIST
	log.Printf("Starting to generate TTS for: %s", data.Title)
	_, err = p.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Audio.mp3", data.EntryID)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				log.Printf("Audio does not exist, generating audio")
			default:
				log.Printf("Failed to check if object exists: %v", err)
				return err
			}
		}
	}
	if err != nil {
		TTSMp3Input := &polly.StartSpeechSynthesisTaskInput{
			OutputFormat:       aws.String("mp3"),
			VoiceId:            aws.String("Joey"),
			Engine:             aws.String("standard"),
			Text:               aws.String(string(summary)),
			OutputS3BucketName: aws.String("lecture-processor"),
			OutputS3KeyPrefix:  aws.String(fmt.Sprintf("%s/Audio-", data.EntryID)),
		}

		mp3Generation, err := p.PollyClient.StartSpeechSynthesisTask(TTSMp3Input)
		if err != nil {
			log.Printf("API to AWS Polly (Audio) Failed: %v", err)
			return err
		}

		// Constantly check if the audio file has been generated since StartSpeechSynthesisTask is asynchronous
		// Poll using a ticker until the task is complete or two minutes have passed
		timeout := time.After(2 * time.Minute)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
	Mp3Loop:
		for {
			select {
			case <-ticker.C:
				taskInput := &polly.GetSpeechSynthesisTaskInput{
					TaskId: mp3Generation.SynthesisTask.TaskId,
				}
				task, err := p.PollyClient.GetSpeechSynthesisTask(taskInput)
				if err != nil {
					log.Printf("Failed to get task status: %v", err)
					return err
				}
				if task.SynthesisTask.TaskStatus != nil && *task.SynthesisTask.TaskStatus == polly.TaskStatusCompleted {
					log.Printf("Generated TTS audio for: %s", data.Title)

					// Once generated, we rename the file to Audio.mp3 and remove the lifecycle policy of 72 hours
					_, err = p.s3Client.CopyObject(&s3.CopyObjectInput{
						Bucket:     aws.String("lecture-processor"),
						CopySource: aws.String(fmt.Sprintf("lecture-processor/%s/Audio-.%s.mp3", data.EntryID, *task.SynthesisTask.TaskId)),
						Key:        aws.String(fmt.Sprintf("%s/Audio.mp3", data.EntryID)),
					})
					if err != nil {
						log.Printf("Failed to copy object (%s): %v", fmt.Sprintf("%s/Audio-.%s.mp3", data.EntryID, *task.SynthesisTask.TaskId), err)
						return err
					}
					_, err = p.s3Client.DeleteObject(&s3.DeleteObjectInput{
						Bucket: aws.String("lecture-processor"),
						Key:    aws.String(fmt.Sprintf("%s/Audio-.%s.mp3", data.EntryID, *task.SynthesisTask.TaskId)),
					})
					if err != nil {
						log.Printf("Failed to delete object (%s): %v", fmt.Sprintf("%s/Audio-.%s.mp3", data.EntryID, *task.SynthesisTask.TaskId), err)
						return err
					}
					break Mp3Loop
				}

			case <-timeout:
				log.Printf("Failed to generate TTS audio for: %s", data.Title)
				return errors.New("failed to generate TTS audio")
			}
		}
	} else {
		log.Printf("Audio already exists for: %s", data.Title)
	}

	// [PRECHECK] : DO NOT REPROCESS IF TTS MARKS ALREADY EXIST
	log.Printf("Starting to generate TTS marks for: %s", data.Title)
	_, err = p.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Words.marks", data.EntryID)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				log.Printf("Marks do not exist, generating marks")
			default:
				log.Printf("Failed to check if object exists: %v", err)
				return err
			}
		}
	}
	if err != nil {
		TTSSubtitlesInput := &polly.StartSpeechSynthesisTaskInput{
			OutputFormat:       aws.String("json"),
			Text:               aws.String(string(summary)),
			VoiceId:            aws.String("Joey"),
			Engine:             aws.String("standard"),
			SpeechMarkTypes:    aws.StringSlice([]string{"word"}),
			OutputS3BucketName: aws.String("lecture-processor"),
			OutputS3KeyPrefix:  aws.String(fmt.Sprintf("%s/Words-", data.EntryID)),
		}

		subtitlesGeneration, err := p.PollyClient.StartSpeechSynthesisTask(TTSSubtitlesInput)
		if err != nil {
			log.Printf("API to AWS Polly (Subtitle) Failed: %v", err)
			return err
		}

		// Constantly check if the json file has been generated since StartSpeechSynthesisTask is asynchronous
		// Poll using a ticker until the task is complete or two minutes have passed
		timeout := time.After(2 * time.Minute)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
	WordsLoop:
		for {
			select {
			case <-ticker.C:
				taskInput := &polly.GetSpeechSynthesisTaskInput{
					TaskId: subtitlesGeneration.SynthesisTask.TaskId,
				}
				task, err := p.PollyClient.GetSpeechSynthesisTask(taskInput)
				if err != nil {
					log.Printf("Failed to get task status: %v", err)
					return err
				}
				if task.SynthesisTask.TaskStatus != nil && *task.SynthesisTask.TaskStatus == polly.TaskStatusCompleted {
					log.Printf("Generated TTS subtitles for: %s", data.Title)

					// Once generated, we rename the file to Words.marks and remove the lifecycle policy of 72 hours
					_, err = p.s3Client.CopyObject(&s3.CopyObjectInput{
						Bucket:     aws.String("lecture-processor"),
						CopySource: aws.String(fmt.Sprintf("lecture-processor/%s/Words-.%s.marks", data.EntryID, *task.SynthesisTask.TaskId)),
						Key:        aws.String(fmt.Sprintf("%s/Words.marks", data.EntryID)),
					})
					if err != nil {
						log.Printf("Failed to copy object (%s): %v", fmt.Sprintf("%s/Words-.%s.marks", data.EntryID, *task.SynthesisTask.TaskId), err)
						return err
					}
					_, err = p.s3Client.DeleteObject(&s3.DeleteObjectInput{
						Bucket: aws.String("lecture-processor"),
						Key:    aws.String(fmt.Sprintf("%s/Words-.%s.marks", data.EntryID, *task.SynthesisTask.TaskId)),
					})
					if err != nil {
						log.Printf("Failed to delete object (%s): %v", fmt.Sprintf("%s/Words-.%s.marks", data.EntryID, *task.SynthesisTask.TaskId), err)
						return err
					}
					break WordsLoop
				}

			case <-timeout:
				log.Printf("Failed to generate TTS subtitles for: %s", data.Title)
				return errors.New("failed to generate TTS subtitles")
			}
		}
	} else {
		log.Printf("Marks already exist for: %s", data.Title)
	}

	// [PRECHECK] : DO NOT REPROCESS IF TTS SUBTITLES ALREADY EXIST
	log.Printf("Starting to generate TTS subtitles for: %s", data.Title)
	_, err = p.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Subtitles.ass", data.EntryID)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				log.Printf("Subtitles do not exist, generating subtitles")
			default:
				log.Printf("Failed to check if object exists: %v", err)
				return err
			}
		}
	}
	if err != nil {
		mp3GetResult, err := p.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String("lecture-processor"),
			Key:    aws.String(fmt.Sprintf("%s/Audio.mp3", data.EntryID)),
		})
		if err != nil {
			log.Printf("Failed to get audio from S3: %v", err)
			return err
		}

		Mp3Fp, err := os.CreateTemp("", "tts-*.mp3")
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			return err
		}
		defer os.Remove(Mp3Fp.Name())
		defer Mp3Fp.Close()

		_, err = io.Copy(Mp3Fp, mp3GetResult.Body)
		if err != nil {
			log.Printf("Failed to write to temp file: %v", err)
			return err
		}
		// Convert the TTS Subtitles to ASS Format
		AssFp, err := os.CreateTemp("", "subtitle-*.ass")
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			return err
		}
		defer os.Remove(AssFp.Name())
		defer AssFp.Close()

		// Get data from S3
		subtitlesGetResult, err := p.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String("lecture-processor"),
			Key:    aws.String(fmt.Sprintf("%s/Words.marks", data.EntryID)),
		})
		if err != nil {
			log.Printf("Failed to get words from S3: %v", err)
			return err
		}

		words, err := ParseJSON(subtitlesGetResult.Body)
		if err != nil {
			log.Printf("Failed to parse subtitles")
			return nil
		}

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

		// Seek to the beginning of the file before uploading
		if _, err := AssFp.Seek(0, io.SeekStart); err != nil {
			log.Printf("Failed to seek to the beginning of the file: %v", err)
			return err
		}

		_, err = p.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String("lecture-processor"),
			Key:         aws.String(fmt.Sprintf("%s/Subtitles.ass", data.EntryID)),
			ContentType: aws.String("application/x-ass"),
			Body:        AssFp,
		})
		if err != nil {
			log.Printf("Failed to upload subtitle file to S3")
			return err
		}

		log.Printf("Generated TTS subtitles for: %s ", data.Title)
	} else {
		log.Printf("Subtitles already exist for: %s", data.Title)
	}

	task, err := NewGenerateVideoTask(&models.GenerateVideoInformation{
		EntryID:         data.EntryID,
		Title:           data.Title,
		BackgroundVideo: data.BackgroundVideo,
	})
	if err != nil {
		log.Printf("Failed to generate Video Task")
		return err
	}
	_, err = p.asynqClient.Enqueue(task)
	if err != nil {
		log.Printf("Failed to queue Video Generation")
		return err
	}

	log.Printf("Finished TTS processing: %s", data.Title)

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
	var lineText []subtitleWord
	var strLen int
	for i, word := range words {
		if i == 0 {
			line.Start = word.Time
			lineText = append(lineText, word)
			strLen = len(word.Value)
		} else {
			if strLen+len(word.Value)+1 <= CHARSPERLINE {
				lineText = append(lineText, word)
				strLen += len(word.Value) + 1
			} else {
				line.End = word.Time - 10
				line.Text = lineText
				lines = append(lines, line)
				line = subtitleLine{}
				line.Start = word.Time + 10
				lineText = []subtitleWord{word}
				strLen = len(word.Value)
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

const HIGHLIGHT_COLOR = "\\1c&HF755A8&"

func GenerateAAS(lines []subtitleLine, aasFptr *os.File) error {
	// Write the Static Script Info Header
	aasFptr.WriteString("[Script Info]\n")
	aasFptr.WriteString("PlayResX: 1080\n")
	aasFptr.WriteString("PlayResY: 1920\n")
	aasFptr.WriteString("WrapStyle: 0\n\n")

	// Write the V4+ Styles Header
	aasFptr.WriteString("[V4+ Styles]\n")
	aasFptr.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	aasFptr.WriteString("Style: Default,Berlin Sans FB,50,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,4,4,2,10,10,10,1\n\n")

	// Write The Events Header
	aasFptr.WriteString("[Events]\n")
	aasFptr.WriteString("Format: Layer, Start, End, Style, Text\n")
	for _, line := range lines {
		aasFptr.WriteString("Dialogue: 0," + line.startAsString() + "," + line.endAsString() + ",Default,{\\an5\\pos(540,960)\\fscx120\\fscy120\\alpha&HFF&\\t(0,35,\\alpha&H00&)\\t(0,35,\\fscx170\\fscy170)\\t(35,75,\\fscx160\\fscy160)\\}")
		for i, word := range line.Text {
			// Format should generate as follows: {\1c&HFFFFFF&\t(start,start,HIGHLIGHT_COLOR)\t(end,end,\1c&HFFFFFF&)}Word
			// If it is the first word then start is 0.
			// If the word is the last word then set the end as the duration (line.End - line.Start). Then move to the next line.

			// Starting the line.
			startOffset := line.Text[0].Time
			if i == 0 {
				// Start at 75 which is after the pop out animations occur.
				aasFptr.WriteString(fmt.Sprintf("{\\1c&HFFFFFF&\\t(75,75,%s)", HIGHLIGHT_COLOR))
			} else {
				prevText := line.Text[i-1].Value
				aasFptr.WriteString(fmt.Sprintf("\\t(%d,%d,\\1c&HFFFFFF&)}%s ", word.Time-startOffset, word.Time-startOffset, prevText))
				aasFptr.WriteString(fmt.Sprintf("{\\1c&HFFFFFF&\\t(%d,%d,%s)", word.Time-startOffset, word.Time-startOffset, HIGHLIGHT_COLOR))
			}

			// Ending the line
			if i == len(line.Text)-1 {
				aasFptr.WriteString(fmt.Sprintf("\\t(%d,%d,\\1c&HFFFFFF&)}%s", line.End-startOffset, line.End-startOffset, word.Value))
			}
		}
		aasFptr.WriteString("\n")
	}
	return nil
}
