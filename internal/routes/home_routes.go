package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterHomeRoutes(api *gin.RouterGroup) {
	homeController := controllers.NewHomeController()
	api.GET("/", homeController.GetHome)
}
