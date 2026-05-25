package utils

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{"success": true, "message": message, "data": data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "message": message})
}

func ValidationError(c *gin.Context, status int, message string, details any) {
	c.JSON(status, gin.H{"success": false, "message": message, "errors": details})
}
