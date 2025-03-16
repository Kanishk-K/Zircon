package subtitleclient

import "fmt"

const MAX_CHARS_PER_LINE = 25
const HIGHLIGHT_COLOR = "\\1c&H639fc5&"

type WordTimeStamp struct {
	Word      string  `json:"word"`
	StartTime float64 `json:"start"`
	EndTime   float64 `json:"end"`
}

type LineTimeStamp struct {
	Line []WordTimeStamp `json:"line"`
}

func (w WordTimeStamp) Duration() (string, string) {
	// Returns the duration of the word in HH:MM:SS.MS format
	// Times are given in seconds
	startDurations := []int{0, 0, 0, 0}
	startTotal := w.StartTime
	if startTotal > 0 {
		startDurations[3] = int(startTotal / (60 * 60)) // Number of seconds in an hour
		startTotal -= float64(startDurations[3]) * 60 * 60
		startDurations[2] = int(startTotal / 60) // Number of seconds in a minute
		startTotal -= float64(startDurations[2]) * 60
		startDurations[1] = int(startTotal)
		startTotal -= float64(startDurations[1])
		startDurations[0] = int(startTotal * 100)
	}
	endDurations := []int{0, 0, 0, 0}
	endTotal := w.EndTime
	if endTotal > 0 {
		endDurations[3] = int(endTotal / (60 * 60)) // Number of seconds in an hour
		endTotal -= float64(endDurations[3]) * 60 * 60
		endDurations[2] = int(endTotal / 60) // Number of seconds in a minute
		endTotal -= float64(endDurations[2]) * 60
		endDurations[1] = int(endTotal)
		endTotal -= float64(endDurations[1])
		endDurations[0] = int(endTotal * 100)
	}
	return fmt.Sprintf("%02d:%02d:%02d.%02d", startDurations[3], startDurations[2], startDurations[1], startDurations[0]), fmt.Sprintf("%02d:%02d:%02d.%02d", endDurations[3], endDurations[2], endDurations[1], endDurations[0])
}

func (l LineTimeStamp) Duration() (string, string) {
	// Returns the start and end time of the line in HH:MM:SS.MS format
	startLine, _ := l.Line[0].Duration()
	_, endLine := l.Line[len(l.Line)-1].Duration()
	return startLine, endLine
}

type LemonFoxResponse struct {
	Audio          string          `json:"audio"`
	WordTimeStamps []WordTimeStamp `json:"word_timestamps"`
}
