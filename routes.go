package main

import (
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/enroll"
	"github.com/s-owl/skhus-backend/grade"
	"github.com/s-owl/skhus-backend/life"
	"github.com/s-owl/skhus-backend/scholarship"
	"github.com/s-owl/skhus-backend/tools"
	"github.com/s-owl/skhus-backend/user"
)

func SetupRoutes(router *gin.Engine) {
	userRoutes := router.Group("/user")
	{
		userRoutes.POST("/login", user.Login)
		userRoutes.GET("/userinfo",
			tools.CredentialOldCheckMiddleware(), user.GetUserinfo)
		userRoutes.GET("/credits",
			tools.CredentialOldCheckMiddleware(), user.GetMyCredits)
		userRoutes.GET("/attendance",
			tools.CredentialOldCheckMiddleware(), user.GetCurrentAttendance)
		userRoutes.POST("/attendance",
			tools.CredentialOldCheckMiddleware(), user.GetAttendanceWithOptions)
		userRoutes.GET("/profile",
			tools.CredentialOldCheckMiddleware(), user.GetUserProfile)
	}

	enrollRoutes := router.Group("/enroll")
	enrollRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		enrollRoutes.GET("/saved_credits", enroll.GetSavedCredits)
		enrollRoutes.GET("/subjects", enroll.GetSubjects)
		enrollRoutes.POST("/subjects", enroll.GetSubjectsWithOptions)
	}

	scholarshipRoutes := router.Group("scholarship")
	scholarshipRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		scholarshipRoutes.GET("history", scholarship.GetScholarshipHistory)
		scholarshipRoutes.GET("result", scholarship.GetScholarshipResults)
	}

	gradeRoutes := router.Group("grade")
	gradeRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		gradeRoutes.GET("certificate", grade.GetGradeCertificate)
	}

	lifeRoutes := router.Group("life")
	{
		lifeRoutes.POST("schedules", life.GetSchedulesWithOptions)
		mealGroup := lifeRoutes.Group("meal")
		{
			mealGroup.GET("urls", life.GetMealURLs)
			mealGroup.POST("data", life.GetMealData)
		}
	}
}
