package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response is a standard API response structure
type Response struct {
	Sucess  bool        `json:"success" example:"true"`
	Status  string      `json:"status" example:"200"`
	Message string      `json:"message" example:"Login successful"`
	Data    interface{} `json:"data,omitempty"`
}

// APIResponse represents a standard JSON response
type APIResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APIUpdateResponse struct {
	Success bool        `json:"success" example:"true"`
	Status  string      `json:"status" example:"200"`
	Message string      `json:"message" example:"User profile updated successfully"`
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

// GetIDParam extracts the "id" parameter from the URL and returns it as an int, or an error if invalid.
func GetIDParam(c *gin.Context, param string) (int, error) {

	idStr := c.Param(param)
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return 0, err
	}
	return id, nil
}