package log

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type iWriter interface {
	Write(in []byte) (n int, err error)
}

type mockWriter struct{}

func (ins *mockWriter) Write(in []byte) (n int, err error) {
	return len(in), nil
}

// enum of logger level
const (
	LevelDebug   int = 3
	LevelInfo    int = 2
	LevelWarning int = 1
	LevelError   int = 0
)

// logger for all log level is exported
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
)

var (
	mockWriterIns *mockWriter
	logPrefix     [4]string
)

// Init initial logger with log level and buffer size
func Init(configLevel int, folderPath string, pidStr string, enableConsole bool, enableFile bool) {
	mockWriterIns = &mockWriter{}
	logPrefix[0] = "ERROR[" + pidStr + "]: "
	logPrefix[1] = "WARN [" + pidStr + "]: "
	logPrefix[2] = "INFO [" + pidStr + "]: "
	logPrefix[3] = "DEBUG[" + pidStr + "]: "

	if enableFile {
		dirErr := os.MkdirAll(folderPath, os.ModePerm)
		if dirErr != nil {
			log.Fatalf("make dir error : %v", dirErr)
		}
	}

	Error = createLogByLevel(filepath.Join(folderPath, "error.log"), LevelError, configLevel, enableConsole, enableFile)
	Warning = createLogByLevel(filepath.Join(folderPath, "warn.log"), LevelWarning, configLevel, enableConsole, enableFile)
	Info = createLogByLevel(filepath.Join(folderPath, "info.log"), LevelInfo, configLevel, enableConsole, enableFile)
	Debug = createLogByLevel(filepath.Join(folderPath, "debug.log"), LevelDebug, configLevel, enableConsole, enableFile)
}

func createLogByLevel(filePath string, level int, configLevel int, enableConsole bool, enableFile bool) *log.Logger {
	var writer iWriter
	if configLevel >= level {

		var fWriter *os.File
		var fErr error

		if enableFile {
			fWriter, fErr = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if fErr != nil {
				log.Fatalf("file open error : %v", fErr)
			}
		}
		if enableConsole {
			if fWriter != nil {
				writer = io.MultiWriter(os.Stdout, fWriter)
			} else {
				writer = os.Stdout
			}
		} else {
			if fWriter != nil {
				writer = fWriter
			} else {
				writer = mockWriterIns
			}
		}
	} else {
		if enableConsole {
			writer = os.Stdout
		} else {
			writer = mockWriterIns
		}
	}
	return log.New(writer, logPrefix[level], log.Ldate|log.Ltime|log.Lshortfile)
}

// Uninit destory logger properly, including flush data and stop schedule
func Uninit() {
	if Error != nil {
		w := Error.Writer()
		if fileWriter, ok := w.(*os.File); ok {
			fileWriter.Sync()
			fileWriter.Close()
		}
	}
	if Warning != nil {
		w := Warning.Writer()
		if fileWriter, ok := w.(*os.File); ok {
			fileWriter.Sync()
			fileWriter.Close()
		}
	}
	if Info != nil {
		w := Info.Writer()
		if fileWriter, ok := w.(*os.File); ok {
			fileWriter.Sync()
			fileWriter.Close()
		}
	}
	if Debug != nil {
		w := Debug.Writer()
		if fileWriter, ok := w.(*os.File); ok {
			fileWriter.Sync()
			fileWriter.Close()
		}
	}
}
