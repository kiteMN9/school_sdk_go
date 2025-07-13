package utils

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	cstZone := time.FixedZone("CST", 8*3600)
	time.Local = cstZone

	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	timeStr := tm.Format("2006-01-02_15-04-05")
	logDir := "logs/"
	if !IsExist(logDir) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			return
		}
	}
	file, err := os.OpenFile(logDir+timeStr+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}

	InitLogger(file, 4096, 800*time.Millisecond)
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

var (
	asyncWriter *asyncLogger
	once        sync.Once
)

type asyncLogger struct {
	bufWriter *bufio.Writer
	lock      sync.Mutex
	flushFreq time.Duration
}

func InitLogger(output *os.File, bufSize int, flushInterval time.Duration) {
	once.Do(func() {

		asyncWriter = &asyncLogger{
			bufWriter: bufio.NewWriterSize(output, bufSize),
			flushFreq: flushInterval,
		}

		log.SetOutput(asyncWriter)

		go asyncWriter.autoFlush()

		go asyncWriter.handleExitSignal()
	})
}

func (a *asyncLogger) Write(p []byte) (n int, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.bufWriter.Write(p)
}

func (a *asyncLogger) autoFlush() {
	for range time.Tick(a.flushFreq) {
		a.lock.Lock()
		err := a.bufWriter.Flush()
		if err != nil {
			a.lock.Unlock()
			return
		}
		a.lock.Unlock()
	}
}

func (a *asyncLogger) handleExitSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	a.lock.Lock()
	err := a.bufWriter.Flush()
	if err != nil {
		a.lock.Unlock()
		return
	}
	a.lock.Unlock()

}
