package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{
			"name":    "Kim JeongTae",
			"version": "0.0.2",
		})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
