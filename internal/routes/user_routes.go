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
	favoriteController := controllers.FavoriteControllerHandler(queries)

	favoriteGroup := api.Group("/favorites")
	favoriteGroup.Use(middlewares.RequireUserAuth())
	{
		favoriteGroup.GET("/", favoriteController.ListFavorites)
		favoriteGroup.POST("/", favoriteController.AddFavorite)

		favoriteGroup.DELETE("/:scholarship_id", favoriteController.RemoveFavorite)
	}

	userGroup := api.Group("/user")
	userGroup.Use(middlewares.RequireUserAuth())
	{
		userGroup.GET("/profile", userController.GetUserProfile)
		userGroup.PATCH("/update-profile", userController.UpdateUserAndProfile)
	}
}
