package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(api *gin.RouterGroup, db *database.Database) {
	queries := importDB.New(db.DB)
	userController := controllers.UserControllerHandler(db.DB, queries)

	protected := api.Group("")
	protected.Use(middlewares.JWTAuthSingle("student"))
	{
		protected.GET("/profile", userController.GetUserProfile)
		protected.PATCH("/update-profile", userController.UpdateUserAndProfile)
	}
}
