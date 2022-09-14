package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var LOG_PATH string

type LogEvents struct {
	mu     sync.Mutex
	Events []*LogEvent `json:"events"`
}

var Logs *LogEvents // global variable used for logging

type LogObject interface {
	LogLine() string
	GetObjType() string
}

type LogEvent struct {
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"endTime"`
	Duration  time.Duration `json:"duration"`
	Desc      string        `json:"description"`
	ObjType   string        `json:"objectType"`
	Obj       LogObject     `json:"object"`
}

// Start and end times should be configured using the pointer returned.
func NewLogEvent(desc string, obj LogObject) *LogEvent {
	return &LogEvent{
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  0,
		Desc:      desc,
		ObjType:   obj.GetObjType(),
		Obj:       obj,
	}
}

func (le *LogEvent) LogLine() string {
	return fmt.Sprintf("%s: %s", FormatTimeNano(le.EndTime), le.Obj.LogLine())
}

func (l *LogEvents) Length() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.Events)
}

func (l *LogEvents) AddLogEvent(event *LogEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Events = append(l.Events, event)
}

func (l *LogEvents) WriteJSON(f io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	bytes, _ := json.MarshalIndent(l, "", "  ")
	f.Write(bytes)
}

type MessageLog struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

func (m *MessageLog) LogLine() string {
	return m.Message
}

func (m *MessageLog) GetObjType() string {
	return "Message"
}

func InitLogger() {
	events := make([]*LogEvent, 0, 1000)
	Logs = &LogEvents{Events: events}

	initStr := "Starting spirit-box..."
	startEvent := NewLogEvent(initStr, &MessageLog{initStr, "spirit-box"})
	startEvent.Duration = startEvent.EndTime.Sub(startEvent.StartTime)
	Logs.AddLogEvent(startEvent)
}

func CreateLogFile() *os.File {
	cur_time := time.Now()
	filename := FormatTime(cur_time) + ".log"

	file, err := os.OpenFile(LOG_PATH+filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close this file somewhere.

	return file
}

func FormatTime(cur_time time.Time) string {
	return fmt.Sprintf(
		"%d-%02d-%02d_%02d:%02d:%02d",
		cur_time.Year(),
		cur_time.Month(),
		cur_time.Day(),
		cur_time.Hour(),
		cur_time.Minute(),
		cur_time.Second(),
	)
}

// Format time with nanosecond precision.
func FormatTimeNano(cur_time time.Time) string {
	return fmt.Sprintf(
		"%d-%02d-%02d_%02d:%02d:%02d.%d",
		cur_time.Year(),
		cur_time.Month(),
		cur_time.Day(),
		cur_time.Hour(),
		cur_time.Minute(),
		cur_time.Second(),
		cur_time.Nanosecond(),
	)
}
