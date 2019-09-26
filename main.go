package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/pprof"
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	SetupRoutes(router)
	pprof.Register(router)
	router.Run()
}
