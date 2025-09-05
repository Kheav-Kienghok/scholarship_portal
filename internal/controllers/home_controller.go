package controllers

import (
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/gin-gonic/gin"
)

// HomeController handles home page requests
type HomeController struct{}

// NewHomeController creates a new home controller instance
func NewHomeController() *HomeController {
	return &HomeController{}
}

// @Summary      Home
// @Description  Returns the home page
// @Tags         home
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1 [get]
func (h *HomeController) GetHome(c *gin.Context) {
	// Log the request
	logging.LogRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP(), http.StatusOK)

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
		"status":  "success",
	})
}
