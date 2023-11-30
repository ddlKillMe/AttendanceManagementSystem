package main

import (
	"dning.com/pro02/controller"
	"dning.com/pro02/middleware"
	"github.com/gin-gonic/gin"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.Use(middleware.CORSMiddleware())
	r.POST("/api/auth/register", controller.Register)
	r.POST("/api/auth/login", controller.Login)
	r.GET("/api/auth/info", middleware.AuthMiddleware(), controller.Info)
	r.POST("/api/auth/leave", controller.Leave)
	r.GET("/api/auth/leaveInfo", controller.LeaveInfo)
	r.GET("/api/auth/deleteLeave", controller.DeleteLeave)

	r.POST("/api/auth/teacherRegister", controller.TeacherRegister)
	r.POST("/api/auth/teacherLogin", controller.TeacherLogin)
	r.GET("/api/auth/teacherInfo", middleware.AuthMiddleware2(), controller.TeacherInfo)
	r.GET("/api/auth/studentInfo", controller.StudentInfo)
	r.GET("/api/auth/studentLeaveInfo", controller.StudentLeaveInfo)
	r.GET("/api/auth/approveLeave", controller.ApproveLeave)
	r.GET("/api/auth/setClassTime", controller.SetClassTime)
	r.GET("/api/auth/signIn", controller.SignIn)
	r.GET("/api/auth/signOut", controller.SignOut)
	r.GET("/api/auth/getAbsence", controller.GetAbsence)
	//r.POST("api/auth/leaveInfo", controller.LeaveInfo)

	return r
}
