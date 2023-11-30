package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Response(c *gin.Context, httpStatus int, code int, data gin.H, msg string) {
	c.JSON(httpStatus, gin.H{"code": code, "data": data, "msg": msg})
}

func Success(c *gin.Context, data gin.H, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "msg": msg})
}

func Fail(c *gin.Context, data gin.H, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 400, "data": data, "msg": msg})
}
