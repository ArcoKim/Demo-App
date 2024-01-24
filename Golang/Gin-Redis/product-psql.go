package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// DBConfig는 데이터베이스 연결 정보를 저장하는 구조체입니다.
type DBConfig struct {
	ReaderEndpoint string `json:"reader_endpoint"`
	WriterEndpoint string `json:"writer_endpoint"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Database       string `json:"database"`
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

	// Reader 엔드포인트를 사용하여 PostgreSQL 데이터베이스에 연결합니다.
	readerDB, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable", config.Username, config.Password, config.Database, config.ReaderEndpoint))
	if err != nil {
		log.Fatal("Failed to connect to the reader database:", err)
	}

	// Writer 엔드포인트를 사용하여 PostgreSQL 데이터베이스에 연결합니다.
	writerDB, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable", config.Username, config.Password, config.Database, config.WriterEndpoint))
	if err != nil {
		log.Fatal("Failed to connect to the writer database:", err)
	}
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

var redisClient *redis.Client

func initRedis() {
	// 환경 변수에서 JSON 형태의 Redis 연결 정보를 가져옵니다.
	redisAddr := os.Getenv("REDIS_ENDPOINT")

	// Redis 클라이언트 초기화
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	// Redis 연결 테스트
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Redis 연결 실패:", err)
	}
}

func main() {
	// Redis 초기화
	initRedis()
	defer redisClient.Close()

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

	// Redis 캐시에 사용자 데이터가 있는지 확인
	cachedUser, err := redisClient.Get(context.Background(), id).Result()
	if err == nil {
		// 캐시에서 사용자 데이터를 찾았으면 반환
		var cachedUserData map[string]string
		json.Unmarshal([]byte(cachedUser), &cachedUserData)
		c.JSON(http.StatusOK, cachedUserData)
		return
	}

	// readerDB를 사용하여 데이터베이스에서 사용자 정보를 조회합니다.
	var name string
	err = readerDB.QueryRow("SELECT name FROM wsi.users WHERE id = $1", id).Scan(&name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// 사용자 데이터를 Redis에 캐시
	userData := map[string]string{"id": id, "name": name}
	jsonData, _ := json.Marshal(userData)
	redisClient.Set(context.Background(), id, jsonData, time.Minute).Result()

	// 사용자 데이터 반환
	c.JSON(http.StatusOK, userData)
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
	_, err := writerDB.Exec("INSERT INTO wsi.users (id, name) VALUES ($1, $2)", user.ID, user.Name)
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
	_, err := writerDB.Exec("UPDATE wsi.users SET name = $1 WHERE id = $2", user.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "name": user.Name})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	// writerDB를 사용하여 데이터베이스에서 사용자 정보를 삭제합니다.
	_, err := writerDB.Exec("DELETE FROM wsi.users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
