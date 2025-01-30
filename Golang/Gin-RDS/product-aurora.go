package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// DBConfig는 데이터베이스 연결 정보를 저장하는 구조체입니다.
type DBConfig struct {
	ReaderEndpoint string `json:"reader_endpoint"`
	WriterEndpoint string `json:"writer_endpoint"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Port           string `json:"port"`
	Database       string `json:"dbname"`
}

var readerDB *sql.DB
var writerDB *sql.DB

func initDB() {
	// 환경 변수에서 JSON 형태의 데이터베이스 연결 정보를 가져옵니다.
	configJSON := os.Getenv("DB_CONFIG_JSON")
	var config DBConfig
	err := json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		log.Fatal("Failed to parse DB config:", err)
	}

	// Reader 엔드포인트를 사용하여 데이터베이스에 연결합니다.
	readerDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Username, config.Password, config.ReaderEndpoint, config.Port, config.Database))
	if err != nil {
		log.Fatal("Failed to connect to the reader database:", err)
	}

	// Writer 엔드포인트를 사용하여 데이터베이스에 연결합니다.
	writerDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Username, config.Password, config.WriterEndpoint, config.Port, config.Database))
	if err != nil {
		log.Fatal("Failed to connect to the writer database:", err)
	}
}

func main() {
	// 데이터베이스 초기화
	initDB()
	defer readerDB.Close()
	defer writerDB.Close()

	// Gin 라우터 생성
	router := gin.Default()

	// API 엔드포인트 설정
	router.GET("/health", healthCheck)
	router.GET("/users/:id", getUser)
	router.POST("/users", createUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)

	// 서버 시작
	router.Run(":8080")
}

func healthCheck(c *gin.Context) {
	// readerDB의 핑 테스트를 수행하여 상태를 확인합니다.
	if err := readerDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func getUser(c *gin.Context) {
	id := c.Param("id")

	// readerDB를 사용하여 데이터베이스에서 사용자 정보를 조회합니다.
	var name string
	err := readerDB.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "name": name})
}

func createUser(c *gin.Context) {
	var user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// writerDB를 사용하여 데이터베이스에 사용자 정보를 추가합니다.
	_, err := writerDB.Exec("INSERT INTO users (id, name) VALUES (?, ?)", user.ID, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "name": user.Name})
}

func updateUser(c *gin.Context) {
	id := c.Param("id")

	var user struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// writerDB를 사용하여 데이터베이스에서 사용자 정보를 업데이트합니다.
	_, err := writerDB.Exec("UPDATE users SET name = ? WHERE id = ?", user.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "name": user.Name})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	// writerDB를 사용하여 데이터베이스에서 사용자 정보를 삭제합니다.
	_, err := writerDB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
