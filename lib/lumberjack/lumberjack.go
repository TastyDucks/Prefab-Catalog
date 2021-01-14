/*
Package lumberjack deals with logging.
Much of this library was inspired [stolen] from Go's default "log" library.
*/
package lumberjack

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

/*
Lumberjack is an active logger object. Each logging operation makes a single call to an io.Writer's Write method. A lumberjack can be used simultaneously from multiple goroutines; it guarantees to serialize access to the Writer.
*/
type Lumberjack struct {
	mx     sync.Mutex // Ensures atomic writes.
	out    io.Writer  // Destination file.
	buf    []byte     // Text buffer.
	prefix string     // Logger prefix.
}

var v int = 0
var p string = ""

/*
Start sets variables for all loggers. It does not return a new Lumberjack object.
*/
func Start(logPath string, verbosity int) {
	v = verbosity
	p = logPath
	err := os.MkdirAll(logPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create log directory: "+err.Error())
		os.Exit(1)
	}
}

/*
New creates and returns a pointer to a new Lumberjack object.
*/
func New(prefix string) *Lumberjack {
	if prefix != "" {
		prefix = fmt.Sprintf("[%s] ", prefix)
	}
	l := &Lumberjack{prefix: prefix}
	return l
}

/*
SetOutput sets the output destination for the Lumberjack.
*/
func (l *Lumberjack) SetOutput(w io.Writer) {
	l.mx.Lock()
	l.out = w
	defer l.mx.Unlock()
}

/*
Log is the main logging function, and requires a logLevel and a log message.
A newline is appended to the log message if one does not exist.
*/
func (l *Lumberjack) Log(logLevel int, text string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	if v > 0 {
		now = time.Now().Format("2006-01-02 15:04:05.000000")
	}
	header := ""
	if logLevel < 0 && v > 0 { // Debug
		header = "Debug"
	}
	if logLevel == 0 && v > -1 { // Info
		header = "Info"
	}
	if logLevel == 1 && v > -2 { // Warn
		header = "Warning"
	}
	if logLevel == 2 && v > -3 { // Error
		header = "Error"
	}
	if logLevel > 2 { // Fatal
		header = "Fatal"
	}
	if header != "" {
		header = fmt.Sprintf("%s\t%s\t[%s]\t", now, l.prefix, header)
		logFileName := time.Now().Format("2006-01-02") + ".log"
		logFile, errOpen := os.OpenFile(p+logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer logFile.Close()
		if errOpen != nil {
			fmt.Fprintf(os.Stderr, "Unable to create log file: %s", errOpen.Error())
			os.Exit(1)
		}
		splitLog := io.MultiWriter(logFile, os.Stdout)
		l.SetOutput(splitLog)
		l.mx.Lock() // Lock access to logger variables.
		l.buf = l.buf[:0]
		l.buf = append(l.buf, header...)
		l.buf = append(l.buf, text...)
		if len(text) == 0 || text[len(text)-1] != '\n' {
			l.buf = append(l.buf, '\n')
		}
		l.mx.Unlock() // Unlock logger.
		_, errWrite := l.out.Write(l.buf)
		if errWrite != nil {
			fmt.Fprintf(os.Stderr, "Unable to write to log file: %s", errWrite.Error())
			os.Exit(1)
		}
	}
}

/*
Debug uses Log(-1, text). Ignored if configuration setting Verbosity < 1.
*/
func (l *Lumberjack) Debug(args ...interface{}) {
	text := fmt.Sprint(args...)
	l.Log(-1, text)
}

/*
Debugf uses Log(-1, text). Ignored if configuration setting Verbosity < 1.
Works like printf. Pass formatted string and any variables after it.
*/
func (l *Lumberjack) Debugf(text string, args ...interface{}) {
	text = fmt.Sprintf(text, args...)
	l.Log(-1, text)
}

/*
Info uses Log(0, text). Ignored if configuration setting Verbosity < 0.
*/
func (l *Lumberjack) Info(args ...interface{}) {
	text := fmt.Sprint(args...)
	l.Log(0, text)
}

/*
Infof uses Log(0, text). Ignored if configuration setting Verbosity < 0.
Works like printf. Pass formatted string and any variables after it.
*/
func (l *Lumberjack) Infof(text string, args ...interface{}) {
	text = fmt.Sprintf(text, args...)
	l.Log(0, text)
}

/*
Warn uses Log(1, text). Ignored if configuration setting Verbosity < -1.
*/
func (l *Lumberjack) Warn(args ...interface{}) {
	text := fmt.Sprint(args...)
	l.Log(1, text)
}

/*
Warnf uses Log(1, text). Ignored if configuration setting Verbosity < -1.
Works like printf. Pass formatted string and any variables after it.
*/
func (l *Lumberjack) Warnf(text string, args ...interface{}) {
	text = fmt.Sprintf(text, args...)
	l.Log(1, text)
}

/*
Error logs an error if it is of type error and is not nil.
Uses Log(2, text). Ignored if configuration setting Verbosity < -2.
*/
func (l *Lumberjack) Error(err error, text string) {
	if err != nil {
		err = errors.Wrap(err, text)
		text = fmt.Sprintf("%s -- %+v", text, err)
		l.Log(2, text)
	}
}

/*
Fatal logs an error if it is of type error and is not nil.
Uses Log(3, text) and runs os.Exit(1).
*/
func (l *Lumberjack) Fatal(err error, text string) {
	if err != nil {
		fmt.Printf("%s", err)
		err = errors.Wrap(err, text)
		text = fmt.Sprintf("%s -- %+v", text, err)
		l.Log(3, text)
		os.Exit(1)
	}
}
