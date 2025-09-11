package routes

import (
	"database/sql"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(api *gin.RouterGroup, db *sql.DB, queries *importDB.Queries) {

	userController := controllers.UserControllerHandler(db, queries)

	protected := api.Group("")
	protected.Use(middlewares.JWTAuthSingle("student"))
	{
		protected.GET("/profile", userController.GetUserProfile)
		protected.PATCH("/update-profile", userController.UpdateUserAndProfile)
	}
}
