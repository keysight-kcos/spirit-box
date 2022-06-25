// Using this as a driver for the development of the logging system.
package main

import (
	"spirit-box/logging"
)

func main() {
	logging.InitLogger()
	l := logging.Logger
	l.Print("test1")
	l.Print("test2")
}
