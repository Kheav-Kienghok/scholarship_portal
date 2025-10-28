package utils

import "github.com/gin-gonic/gin"

// Standard response helpers

func RespondUnauthorized(c *gin.Context, message string) {
	JSONIndent(c, 401, message, nil)
}

func RespondBadRequest(c *gin.Context, message string, data interface{}) {
	JSONIndent(c, 400, message, data)
}

func RespondInternalError(c *gin.Context, message string) {
	JSONIndent(c, 500, message, nil)
}

func RespondOK(c *gin.Context, message string, data interface{}) {
	JSONIndent(c, 200, message, data)
}

func RespondMissingParameter(c *gin.Context, paramName string) {
	msg := "Missing parameter: " + paramName
	JSONIndent(c, 400, msg, nil)
}

func RespondInvalidToken(c *gin.Context) {
	JSONIndent(c, 401, "Invalid token", nil)
}

func RespondInvalidOTP(c *gin.Context) {
	JSONIndent(c, 401, "Invalid OTP", nil)
}

func RespondNotFound(c *gin.Context, entity string) {
	JSONIndent(c, 404, entity+" not found", nil)
}

func RespondTooManyRequests(c *gin.Context, message string, retryAfter int) {
	JSONIndent(c, 429, message, nil)
}
