package main

import (
	"github.com/gin-gonic/gin"
	"github.com/skhus/skhus-backend/users"
)

func SetupUsersRoutes(router *gin.Engine) {
	r := router.Group("/users")
	{
		r.POST("/login", users.Login)
		// r.GET("/userinfo")
		// r.GET("/credits")
		// r.GET("/attendance")
	}
}
