package logging

import (
	"log"
	"fmt"
	"os"
	"time"
)

// All packages will write to this logger after it has been initialized.
// It has built-in serialization for access to its Writer, aka concurrent calls to Print etc are okay.
var Logger *log.Logger

const (
	// LOG_PATH = "/usr/share/spirit-box/logs/"
	LOG_PATH = "./temp_logs/" // temporarily want to work with files in local dir
)

func InitLogger() {
	cur_time := time.Now()
	filename := fmt.Sprintf("spirit-box_%d-%02d-%02d_%02d:%02d:%02d.log",
		cur_time.Year(), cur_time.Month(), cur_time.Day(), cur_time.Hour(), cur_time.Minute(), cur_time.Second())
	fmt.Println(filename)
	file, err := os.OpenFile(LOG_PATH+filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close this file somewhere.

	Logger = log.New(file, "", log.LstdFlags)
}

func Blank() {}
