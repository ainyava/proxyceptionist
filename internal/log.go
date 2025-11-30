package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry represents a single HTTP request log
type LogEntry struct {
	Timestamp   string              `json:"timestamp"`
	Method      string              `json:"method"`
	URL         string              `json:"url"`
	Headers     map[string][]string `json:"headers"`
	RemoteAddr  string              `json:"remote_addr"`
	ProxyTarget string              `json:"proxy_target"`
}

var LogFile *os.File

func InitLog() {
	var err error
	LogFile, err = os.OpenFile("access.json.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
}

// logRequest logs request info in JSON to stdout and file
func LogRequest(c *gin.Context, target string) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Method:      c.Request.Method,
		URL:         c.Request.URL.String(),
		Headers:     c.Request.Header,
		RemoteAddr:  c.ClientIP(),
		ProxyTarget: target,
	}

	data, _ := json.Marshal(entry)
	fmt.Println(string(data))   // stdout
	LogFile.Write(data)         // file
	LogFile.Write([]byte("\n")) // newline
}
