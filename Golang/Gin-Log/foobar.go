package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// 로그 디렉토리 생성
	logDir := "log"
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	// Gin 엔진 초기화
	router := gin.Default()

	// 로그 파일 생성
	logFile, err := os.Create(filepath.Join(logDir, "app.log"))
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// 미들웨어 설정: Access 로그를 log/app.log 파일에 기록
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		clientIP := param.ClientIP
		timestamp := param.TimeStamp.Format(time.RFC3339)
		method := param.Method
		path := param.Path
		protocol := param.Request.Proto
		statusCode := param.StatusCode
		latency := param.Latency
		userAgent := param.Request.UserAgent()

		logFormat := fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" \"\n",
			clientIP, timestamp, method, path, protocol, statusCode, latency, userAgent)

		return logFormat
	}))

	// 라우트 설정
	v1 := router.Group("/v1")
	{
		v1.GET("/foo", func(c *gin.Context) {
			c.JSON(200, gin.H{"application": "foo"})
		})

		v1.GET("/bar", func(c *gin.Context) {
			c.JSON(200, gin.H{"application": "bar"})
		})
	}

	// Healthcheck 엔드포인트
	router.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 서버 시작
	port := ":8080"
	fmt.Printf("Server is running on port %s...\n", port)
	router.Run(port)
}

