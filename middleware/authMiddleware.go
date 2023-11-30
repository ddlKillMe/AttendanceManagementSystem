package middleware

import (
	"dning.com/pro02/common"
	"dning.com/pro02/model"
	"dning.com/pro02/response"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取authorization header
		tokenString := c.GetHeader("Authorization")

		// validate token format
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("token为空或者token前缀不对")
			return
		}

		tokenString = tokenString[7:]

		token, claims, err := common.ParseToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("token无效或者验证token出错")
			return
		}

		SID := claims.SID
		//DB := common.GetDB()
		//var user model.User
		//DB.First(&user, userID)
		// 获取数据库
		rds := common.GetRDS()
		// 获取数据库中的sid
		var student model.Student
		stringMap, err2 := redis.StringMap(rds.Do("HGETALL", redis.Args{}.Add("student:"+SID)...))
		if err2 != nil {
			response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
			log.Printf("中间件获取用户密码时连接redis出错，err: " + err2.Error())
		}
		if stringMap == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("用户被删除")
			return
		}
		student.Name = stringMap["name"]
		student.SID = stringMap["sid"]
		student.Password = stringMap["password"]
		student.NormalCount = stringMap["normalCount"]
		student.LeaveCount = stringMap["leaveCount"]
		student.LateCount = stringMap["lateCount"]
		student.EarlyCount = stringMap["earlyCount"]
		student.AbsenceCount = stringMap["absenceCount"]

		//if user.ID == 0 {
		//	c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
		//	c.Abort()
		//	log.Println("用户被删除")
		//	return
		//}

		c.Set("student", student)
		c.Next()
	}
}

func AuthMiddleware2() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取authorization header
		tokenString := c.GetHeader("Authorization")

		// validate token format
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("token为空或者token前缀不对")
			return
		}

		tokenString = tokenString[7:]

		token, claims, err := common.ParseToken2(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("token无效或者验证token出错")
			return
		}

		TID := claims.TID
		//DB := common.GetDB()
		//var user model.User
		//DB.First(&user, userID)
		// 获取数据库
		rds := common.GetRDS()
		// 获取数据库中的sid
		var teacher model.Teacher
		stringMap, err2 := redis.StringMap(rds.Do("HGETALL", redis.Args{}.Add("teacher:"+TID)...))
		if err2 != nil {
			response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
			log.Printf("中间件获取用户密码时连接redis出错，err: " + err2.Error())
		}
		if stringMap == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			c.Abort()
			log.Println("用户被删除")
			return
		}
		teacher.Name = stringMap["name"]
		teacher.TID = stringMap["tid"]
		teacher.Password = stringMap["password"]

		//if user.ID == 0 {
		//	c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
		//	c.Abort()
		//	log.Println("用户被删除")
		//	return
		//}

		c.Set("teacher", teacher)
		c.Next()
	}
}
