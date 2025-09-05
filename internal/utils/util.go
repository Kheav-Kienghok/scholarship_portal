package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is a standard API response structure
type Response struct {
	Sucess  bool        `json:"success"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONIndent sends a pretty-printed JSON response using the Response struct
func JSONIndent(c *gin.Context, status int, message string, data interface{}) {

	resp := Response{
		Sucess:  status >= 200 && status < 300,
		Status:  http.StatusText(status),
		Message: message,
		Data:    data,
	}

	c.IndentedJSON(status, resp)
}
