package logging

import (
	 "encoding/json"
	 "log"
	"fmt"
	 "io"
	 "os"
	"sync"
	"time"
)

type LogObject interface {
	LogLine() string
}

type LogEvent struct {
	StartTime time.Time
	EndTime time.Time
	Desc string
	Obj LogObject
}

func NewLogEvent(desc string, obj LogObject) *LogEvent {
	return &LogEvent{
		StartTime: time.Now(),
		EndTime: time.Now(),
		Desc: desc,
		Obj: obj,
	}
}

func (le *LogEvent) LogLine() string {
	return fmt.Sprintf("%s: %s", FormatTimeNano(le.StartTime), le.Obj.LogLine())
}

type LogEvents struct {
	mu     sync.Mutex
	Events []*LogEvent
}

var Logs *LogEvents

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

const (
	LOG_PATH = "/usr/share/spirit-box/logs/"
	// LOG_PATH = "/home/severian/data-driven-boot-up-ui/temp_logs/" // temporarily want to work with files in local dir
)

type MessageLog struct {
	Message string
}

func (m *MessageLog) LogLine() string {
	return m.Message
}

func InitLogger() {
	events := make([]*LogEvent, 0, 1000)
	Logs =  &LogEvents{Events: events}

	initStr := "Starting spirit-box..."
	Logs.AddLogEvent(&LogEvent{
		StartTime: time.Now(),
		EndTime: time.Now(),
		Desc: initStr,
		Obj: &MessageLog{initStr},
	})
	/*
	log.Printf("Length of Logs now: %d", Logs.Length())
	log.Print(Logs.Events[0].LogLine())
	*/
}

func CreateLogFile() *os.File {
	cur_time := time.Now()
	filename := FormatTime(cur_time)+".log"

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
