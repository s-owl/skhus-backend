package main

import (
	"context"
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.POST("/user/login", func(c *gin.Context) {
		// create chrome instance
		ctx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
	})
	router.Run()
}
