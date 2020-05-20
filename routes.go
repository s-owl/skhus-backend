package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/enroll"
	"github.com/s-owl/skhus-backend/grade"
	"github.com/s-owl/skhus-backend/life"
	"github.com/s-owl/skhus-backend/scholarship"
	"github.com/s-owl/skhus-backend/tools"
	"github.com/s-owl/skhus-backend/user"
)

// 전체적인 타우팅을 여기에서 구성한다.
func SetupRoutes(router *gin.Engine) {
	// 외부에서 사용하게 만드는 cors 설정, 필용한 곳에만!
	otherConfig := cors.DefaultConfig()
	otherConfig.AllowAllOrigins = true
	otherConfig.AllowHeaders = []string{"Content-Type"}
	accessFromOther := cors.New(otherConfig)
	// 지정한 페이지에서만 사용가능하게 설정
	webConfig := cors.DefaultConfig()
	webConfig.AllowOrigins = consts.SkhusWebSite()
	webConfig.AllowCredentials = true
	webConfig.AllowHeaders = []string{"Content-Type"}
	accessFromWeb := cors.New(webConfig)

	userRoutes := router.Group("/user")
	// 학생과 관련된 정보를 받아오는 API
	userRoutes.Use(accessFromWeb)
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
		userRoutes.GET("/classinfo",
			tools.CredentialOldCheckMiddleware(), user.GetClassInfo)
		userRoutes.GET("/otp",
			tools.CredentialOldCheckMiddleware(), user.GetOtpCode)
		userRoutes.GET("/profile",
			tools.CredentialOldCheckMiddleware(), user.GetUserProfile)
	}

	enrollRoutes := router.Group("/enroll")
	enrollRoutes.Use(accessFromWeb)
	enrollRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		enrollRoutes.GET("/saved_credits", enroll.GetSavedCredits)
		enrollRoutes.GET("/subjects", enroll.GetSubjects)
		enrollRoutes.POST("/subjects", enroll.GetSubjectsWithOptions)
	}

	// 장학에 관련된 정보를 가져오는 API
	scholarshipRoutes := router.Group("scholarship")
	scholarshipRoutes.Use(accessFromWeb)
	scholarshipRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		scholarshipRoutes.GET("history", scholarship.GetScholarshipHistory)
		scholarshipRoutes.GET("result", scholarship.GetScholarshipResults)
	}

	// 학점에 대한 정보를 가져오는 API
	gradeRoutes := router.Group("grade")
	gradeRoutes.Use(accessFromWeb)
	gradeRoutes.Use(tools.CredentialOldCheckMiddleware())
	{
		gradeRoutes.GET("certificate", grade.GetGradeCertificate)
	}

	// 학교 생활(일정, 학식)에 대한 정보를 가져오는 API
	lifeRoutes := router.Group("life")
	lifeRoutes.Use(accessFromOther)
	{
		lifeRoutes.POST("schedules", life.GetSchedulesWithOptions)
		mealGroup := lifeRoutes.Group("meal")
		{
			mealGroup.GET("urls", life.GetMealURLs)
			mealGroup.POST("data", life.GetMealData)
		}
	}
}
