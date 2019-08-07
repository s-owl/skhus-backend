package main

import (
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/user"
)

func SetupUserRoutes(router *gin.Engine) {
	r := router.Group("/user")
	{
		r.POST("/login", user.Login)
		r.GET("/userinfo", user.GetUserinfo)
		// r.GET("/credits")
		// r.GET("/attendance")
	}
}
