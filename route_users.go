package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sukso96100/skhus-backend/users"
)

func SetupUsersRoutes(router *gin.Engine) {
	users := router.Group("/users"){
		users.POST("/login", users.Login)
		users.GET("/userinfo")
		users.GET("/credits")
		users.GET("/attendance")
	}
}
