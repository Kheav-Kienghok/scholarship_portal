package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterScholarshipRoutes(api *gin.RouterGroup, queries *importDB.Queries) {
	scholarshipController := controllers.ScholarshipControllerHandler(queries)

	// Public
	// api.GET("/scholarships", scholarshipController.GetScholarships)

	scholarshipGroup := api.Group("/scholarships")
	{
		// Apply ETag middleware only to GET endpoints
		scholarshipGroup.GET("", middlewares.ETagMiddleware(), scholarshipController.GetScholarships)
		scholarshipGroup.GET("/active", middlewares.ETagMiddleware(), scholarshipController.GetActiveScholarship)
		
		scholarshipGroup.GET("/filter", scholarshipController.FilterByCategory)

		scholarshipGroup.GET("/:id", scholarshipController.GetScholarshipByID)
	}
	
	// Admin only
	admin := api.Group("/scholarships")
	admin.Use(middlewares.RequireAdminAuth())
	{
		// admin.GET("/:id", scholarshipController.GetScholarshipByID)

		admin.POST("", scholarshipController.CreateScholarship)
		admin.DELETE("/:id", scholarshipController.DeleteScholarship)
		
		admin.PATCH("/:id", scholarshipController.UpdateScholarship)
	}
	
	// User authenticated
	user := api.Group("/scholarships")
	user.Use(middlewares.RequireUserAuth())
	{
		user.GET("/search", scholarshipController.SearchScholarships)
	}
}
