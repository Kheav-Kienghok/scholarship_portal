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
	api.GET("/scholarships", scholarshipController.GetScholarships)

	// Admin only
	admin := api.Group("/scholarships")
	admin.Use(middlewares.JWTAuthSingle("admin"))
	{
		admin.POST("", scholarshipController.CreateScholarship)
		// admin.GET("/:id", scholarshipController.GetScholarshipByID)
		// admin.PUT("/:id", scholarshipController.UpdateScholarship)
		// admin.DELETE("/:id", scholarshipController.DeleteScholarship)
	}
}
