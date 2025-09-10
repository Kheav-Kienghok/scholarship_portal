package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/gin-gonic/gin"
)

func RegisterHomeRoutes(api *gin.RouterGroup, db *database.Database) {
	homeController := controllers.NewHomeController()
	api.GET("/", homeController.GetHome)
}
