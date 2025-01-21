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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
	"k8s.io/apimachinery/pkg/util/wait"
)

/* JOB SETUP */
const TypeGenerateVideo = "summary:genVideo"

type GenerateVideoProcess struct {
	pollyClient  *polly.Polly
	s3Client     *s3.S3
	dynamoClient services.DynamoMethods
}

func NewGenerateVideoProcess(pollyClient *polly.Polly, s3Client *s3.S3, dynamoClient services.DynamoMethods) *GenerateVideoProcess {
	return &GenerateVideoProcess{
		pollyClient:  pollyClient,
		s3Client:     s3Client,
		dynamoClient: dynamoClient,
	}
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
const HIGHLIGHT_COLOR = "\\1c&HF755A8&"

/* Producer Call */

func NewGenerateVideoTask(jobInfo *models.GenerateVideoInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateVideo, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
}

/* Consumer Call */

func (p *GenerateVideoProcess) HandleGenerateVideoTask(ctx context.Context, t *asynq.Task) error {
	// What does our task require?
	data := models.GenerateVideoInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		return err
	}

	log.Printf("Processing video generation for entry: %s", data.EntryID)

	workingDir, err := os.MkdirTemp("", data.EntryID)
	if err != nil {
		log.Printf("Error in creating temp directory: %v", err)
		return err
	}
	defer os.RemoveAll(workingDir)

	mp3Fp, err := os.CreateTemp(workingDir, "tts-*.mp3")
	if err != nil {
		log.Printf("Failed to provision a temporary mp3 file: %v", err)
		return err
	}
	defer mp3Fp.Close()
	defer os.Remove(mp3Fp.Name())

	subtitlesFp, err := os.CreateTemp(workingDir, "tts-*.ass")
	if err != nil {
		log.Printf("Failed to provision a temporary ass file: %v", err)
		return err
	}
	defer subtitlesFp.Close()
	defer os.Remove(subtitlesFp.Name())

	// We are guarenteed to have a Summary.
	// We may or may NOT have a mp3 AND marks.
	// If we do, download and generate video. Otherwise generate both files.

	if data.GenerateSubtitles {
		// Generate the items and upload
		log.Printf("Failed to find MP3 or Subtitles for %s", data.EntryID)
		err = p.GenerateMediaFiles(ctx, mp3Fp, subtitlesFp, &data.EntryID)
		if err != nil {
			log.Printf("Failed to generate media files: %v", err)
			return err
		}
		// Update Database and Redis
		p.dynamoClient.UpdateSubtitles(data.EntryID)
	} else {
		log.Printf("Found both MP3 or Subtitles for %s", data.EntryID)
		err = p.FetchMediaFiles(ctx, mp3Fp, subtitlesFp, &data.EntryID)
		if err != nil {
			log.Printf("Failed to fetch media files: %v", err)
			return err
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to get current working directory.")
		return err
	}
	backgroundVideo := filepath.Join(dir, "static", fmt.Sprintf("%s.mp4", data.BackgroundVideo))
	cmd := exec.Command("ffmpeg", "-y", "-i", backgroundVideo, "-i", filepath.Base(mp3Fp.Name()), "-vf", fmt.Sprintf("subtitles='%s'", filepath.Base(subtitlesFp.Name())), "-filter:a", fmt.Sprintf("atempo=%.2f", TEMPOSPEED), "-c:v", "libx264", "-c:a", "aac", "-crf", "30", "-shortest", "output.mp4")
	cmd.Dir = workingDir
	log.Printf("Generating video for %s", data.EntryID)
	err = cmd.Run()
	if err != nil {
		log.Printf("Error in running ffmpeg command: %v", err)
		return err
	}

	outputFp, err := os.Open(filepath.Join(workingDir, "output.mp4"))
	if err != nil {
		log.Printf("Failed to open output Mp4 file: %v", err)
		return err
	}
	defer outputFp.Close()
	defer os.Remove(outputFp.Name())

	// Upload the video to S3
	_, err = p.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("lecture-processor"),
		Key:         aws.String(fmt.Sprintf("assets/%s/%s.mp4", data.EntryID, data.BackgroundVideo)),
		ContentType: aws.String("video/mp4"),
		Body:        outputFp,
	})

	if err != nil {
		log.Printf("Failed to upload video to S3: %v", err)
		return err
	}

	err = p.dynamoClient.AddVideo(data.EntryID, data.BackgroundVideo)
	if err != nil {
		log.Printf("Failed to update job status on DynamoDB: %v", err)
		return err
	}

	log.Printf("Completed video generation for entry: %s", data.EntryID)

	return nil
}

func (p *GenerateVideoProcess) GenerateMediaFiles(ctx context.Context, mp3Fp *os.File, subtitlesFp *os.File, entryID *string) error {
	// Get the summary
	result, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Summary.txt", *entryID)),
	})
	if err != nil {
		log.Printf("Failed to get summary from S3: %v", err)
		return err
	}
	defer result.Body.Close()

	summary, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Failed to copy summary into buffer: %v", err)
		return err
	}

	// Generate and fill up mp3Fp.
	duration, err := p.GenerateMP3(ctx, mp3Fp, &summary, entryID)
	if err != nil {
		log.Printf("Failed to generate MP3 file: %v", err)
		return err
	}
	err = p.GenerateSubtitles(ctx, subtitlesFp, &summary, entryID, duration)
	if err != nil {
		log.Printf("Failed to generate ASS file: %v", err)
		return err
	}

	return nil
}

func (p *GenerateVideoProcess) FetchMediaFiles(ctx context.Context, mp3Fp *os.File, subtitlesFp *os.File, entryID *string) error {
	// Get MP3
	mp3Output, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Audio.mp3", *entryID)),
	})
	if err != nil {
		log.Printf("Failed to fetch Mp3 from S3: %v", err)
		return err
	}
	defer mp3Output.Body.Close()
	_, err = io.Copy(mp3Fp, mp3Output.Body)
	if err != nil {
		log.Printf("Failed to copy Mp3 contents to temp file: %v", err)
		return err
	}

	// Get Subtitles
	subtitlesOutput, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Subtitles.ass", *entryID)),
	})
	if err != nil {
		log.Printf("Failed to fetch Subtitle from S3: %v", err)
		return err
	}
	defer subtitlesOutput.Body.Close()
	_, err = io.Copy(subtitlesFp, subtitlesOutput.Body)
	if err != nil {
		log.Printf("Failed to copy Subtitle contents to temp file: %v", err)
		return err
	}

	return nil
}

func (p *GenerateVideoProcess) awaitGeneration(jobOutput *polly.GetSpeechSynthesisTaskInput) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		check, err := p.pollyClient.GetSpeechSynthesisTask(jobOutput)
		if err != nil {
			log.Printf("Failed to get task completion information: %v", err)
			return false, err
		}
		return (check.SynthesisTask.TaskStatus != nil && *check.SynthesisTask.TaskStatus == polly.TaskStatusCompleted), nil
	}
}

func (p *GenerateVideoProcess) GenerateMP3(ctx context.Context, mp3Fp *os.File, summary *[]byte, entryID *string) (int, error) {
	TTSMp3Input := &polly.StartSpeechSynthesisTaskInput{
		OutputFormat:       aws.String(polly.OutputFormatMp3),
		VoiceId:            aws.String(polly.VoiceIdJoey),
		Engine:             aws.String(polly.EngineStandard),
		Text:               aws.String(string(*summary)),
		OutputS3BucketName: aws.String("lecture-processor"),
		OutputS3KeyPrefix:  aws.String(fmt.Sprintf("assets/%s/Audio-", *entryID)),
	}

	mp3Generation, err := p.pollyClient.StartSpeechSynthesisTask(TTSMp3Input)
	if err != nil {
		log.Printf("Failed to generate MP3 with Polly: %v", err)
		return -1, err
	}
	mp3Monitor := &polly.GetSpeechSynthesisTaskInput{
		TaskId: mp3Generation.SynthesisTask.TaskId,
	}
	err = wait.PollUntilContextTimeout(ctx, time.Second, 5*time.Minute, true, p.awaitGeneration(mp3Monitor))
	if err != nil {
		log.Printf("Failed to poll successfully for Mp3: %v", err)
		return -1, err
	}

	_, err = p.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String("lecture-processor"),
		CopySource: aws.String(fmt.Sprintf("lecture-processor/assets/%s/Audio-.%s.mp3", *entryID, *mp3Generation.SynthesisTask.TaskId)),
		Key:        aws.String(fmt.Sprintf("assets/%s/Audio.mp3", *entryID)),
	})
	if err != nil {
		log.Printf("Failed to copy mp3 file: %v", err)
		return -1, err
	}
	_, err = p.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Audio-.%s.mp3", *entryID, *mp3Generation.SynthesisTask.TaskId)),
	})
	if err != nil {
		log.Printf("Failed to delete temp mp3 file: %v", err)
		return -1, err
	}
	output, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Audio.mp3", *entryID)),
	})
	if err != nil {
		log.Printf("Failed to download mp3 file from s3: %v", err)
		return -1, err
	}
	defer output.Body.Close()
	_, err = io.Copy(mp3Fp, output.Body)
	if err != nil {
		log.Printf("Failed to copy S3 MP3 data to tempfile: %v", err)
		return -1, err
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", mp3Fp.Name())
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get the duration of the mp3 file: %v", err)
		return -1, err
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

func (p *GenerateVideoProcess) GenerateSubtitles(ctx context.Context, subtitlesFp *os.File, summary *[]byte, entryID *string, duration int) error {
	TTSMarksInput := &polly.StartSpeechSynthesisTaskInput{
		OutputFormat:       aws.String(polly.OutputFormatJson),
		VoiceId:            aws.String(polly.VoiceIdJoey),
		Engine:             aws.String(polly.EngineStandard),
		Text:               aws.String(string(*summary)),
		SpeechMarkTypes:    aws.StringSlice([]string{polly.SpeechMarkTypeWord}),
		OutputS3BucketName: aws.String("lecture-processor"),
		OutputS3KeyPrefix:  aws.String(fmt.Sprintf("assets/%s/Words-", *entryID)),
	}

	marksGeneration, err := p.pollyClient.StartSpeechSynthesisTask(TTSMarksInput)
	if err != nil {
		log.Printf("Failed to generate Marks with Polly: %v", err)
		return err
	}
	marksMonitor := &polly.GetSpeechSynthesisTaskInput{
		TaskId: marksGeneration.SynthesisTask.TaskId,
	}
	err = wait.PollUntilContextTimeout(ctx, time.Second, 5*time.Minute, true, p.awaitGeneration(marksMonitor))
	if err != nil {
		log.Printf("Failed to poll successfully for Marks: %v", err)
		return err
	}

	output, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("assets/%s/Words-.%s.marks", *entryID, *marksGeneration.SynthesisTask.TaskId)),
	})
	if err != nil {
		log.Printf("Failed to download mp3 file from s3: %v", err)
		return err
	}
	defer output.Body.Close()

	subtitleWords, err := ParseJSON(output.Body)
	if err != nil {
		log.Printf("Failed to parse words from JSON: %v", err)
		return err
	}
	subtitleLines := GenerateSubtitleLines(subtitleWords, duration)
	err = GenerateAAS(subtitleLines, subtitlesFp)
	if err != nil {
		log.Printf("Failed to generate subtitle file: %v", err)
		return err
	}

	_, err = p.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("lecture-processor"),
		Key:         aws.String(fmt.Sprintf("assets/%s/Subtitles.ass", *entryID)),
		ContentType: aws.String("application/x-ass"),
		Body:        subtitlesFp,
	})
	if err != nil {
		log.Printf("Failed to upload subtitle file to S3: %v", err)
		return err
	}

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

func GenerateAAS(lines []subtitleLine, assFptr *os.File) error {
	// Write the Static Script Info Header
	assFptr.WriteString("[Script Info]\n")
	assFptr.WriteString("PlayResX: 1080\n")
	assFptr.WriteString("PlayResY: 1920\n")
	assFptr.WriteString("WrapStyle: 0\n\n")

	// Write the V4+ Styles Header
	assFptr.WriteString("[V4+ Styles]\n")
	assFptr.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	assFptr.WriteString("Style: Default,Berlin Sans FB,50,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,4,4,2,10,10,10,1\n\n")

	// Write The Events Header
	assFptr.WriteString("[Events]\n")
	assFptr.WriteString("Format: Layer, Start, End, Style, Text\n")
	for _, line := range lines {
		assFptr.WriteString("Dialogue: 0," + line.startAsString() + "," + line.endAsString() + ",Default,{\\an5\\pos(540,960)\\fscx120\\fscy120\\alpha&HFF&\\t(0,35,\\alpha&H00&)\\t(0,35,\\fscx170\\fscy170)\\t(35,75,\\fscx160\\fscy160)\\}")
		for i, word := range line.Text {
			// Format should generate as follows: {\1c&HFFFFFF&\t(start,start,HIGHLIGHT_COLOR)\t(end,end,\1c&HFFFFFF&)}Word
			// If it is the first word then start is 0.
			// If the word is the last word then set the end as the duration (line.End - line.Start). Then move to the next line.

			// Starting the line.
			startOffset := line.Text[0].Time
			if i == 0 {
				// Start at 75 which is after the pop out animations occur.
				assFptr.WriteString(fmt.Sprintf("{\\1c&HFFFFFF&\\t(75,75,%s)", HIGHLIGHT_COLOR))
			} else {
				prevText := line.Text[i-1].Value
				assFptr.WriteString(fmt.Sprintf("\\t(%d,%d,\\1c&HFFFFFF&)}%s ", word.Time-startOffset, word.Time-startOffset, prevText))
				assFptr.WriteString(fmt.Sprintf("{\\1c&HFFFFFF&\\t(%d,%d,%s)", word.Time-startOffset, word.Time-startOffset, HIGHLIGHT_COLOR))
			}

			// Ending the line
			if i == len(line.Text)-1 {
				assFptr.WriteString(fmt.Sprintf("\\t(%d,%d,\\1c&HFFFFFF&)}%s", line.End-startOffset, line.End-startOffset, word.Value))
			}
		}
		assFptr.WriteString("\n")
	}
	if _, err := assFptr.Seek(0, io.SeekStart); err != nil {
		log.Printf("Failed to seek to the beginning of the file: %v", err)
		return err
	}
	return nil
}
