package main

import (
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/enroll"
	"github.com/s-owl/skhus-backend/scholarship"
	"github.com/s-owl/skhus-backend/user"
)

func SetupRoutes(router *gin.Engine) {
	userRoutes := router.Group("/user")
	{
		userRoutes.POST("/login", user.Login)
		userRoutes.GET("/userinfo", user.GetUserinfo)
		userRoutes.GET("/credits", user.GetMyCredits)
		userRoutes.GET("/attendance", user.GetCurrentAttendance)
		userRoutes.POST("/attendance", user.GetAttendanceWithOptions)
	}

	enrollRoutes := router.Group("/enroll")
	{
		enrollRoutes.GET("/saved_credits", enroll.GetSavedCredits)
		enrollRoutes.GET("/subjects", enroll.GetSubjects)
		enrollRoutes.POST("/subjects", enroll.GetSubjectsWithOptions)
	}

	scholarshipRoutes := router.Group("scholarship")
	{
		scholarshipRoutes.GET("history", scholarship.GetScholarshipHistory)
		scholarshipRoutes.GET("result", scholarship.GetScholarshipResults)
	}
}
