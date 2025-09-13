package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BindJSONOrFail binds JSON to a struct and handles missing/invalid fields automatically.
// Usage: if !BindJSONOrFail(c, &input) { return }
func BindJSONOrFail(c *gin.Context, obj interface{}) bool {

	if err := c.ShouldBindJSON(obj); err != nil {

		if errs, ok := err.(validator.ValidationErrors); ok {
			field := errs[0].Field()
			RespondMissingParameter(c, field)
			return false
		}

		RespondBadRequest(c, "Invalid input", err.Error())
		return false
	}
	return true
}
