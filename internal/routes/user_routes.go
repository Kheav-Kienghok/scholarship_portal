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

	userGroup := api.Group("/user")
	userGroup.Use(middlewares.RequireUserAuth())
	{
		userGroup.GET("/profile", userController.GetProfile)
		userGroup.PUT("/profile", userController.UpdateProfile)
		// userGroup.PATCH("/update-profile", userController.UpdateUserAndProfile)

		// User favorites
		userGroup.GET("/favorites", favoriteController.ListFavorites)
	}

	// Top-level favorite routes for POST / DELETE
	favoriteGroup := api.Group("/favorite")
	favoriteGroup.Use(middlewares.RequireUserAuth())
	{
		favoriteGroup.POST("", favoriteController.AddFavorite)
		favoriteGroup.PUT("/:scholarship_id", favoriteController.UpdateFavoriteStatus)

		favoriteGroup.DELETE("/:scholarship_id", favoriteController.RemoveFavorite)

	}
}
