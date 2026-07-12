package controllers

import (
	"net/http"
	"yourtube/internal/models"
	"yourtube/internal/repositories"

	"github.com/gin-gonic/gin"
)

func AddUser(ctx *gin.Context) {
	db := repositories.GetDb()

	type AddUserRequestBody struct {
		Name     string                  `json:"name"`
		Email    string                  `json:"email"`
		FamilyId int32                   `json:"familyId"`
		Role     models.FamilyMemberRole `json:"role"`
	}
	var body AddUserRequestBody
	ctx.ShouldBindJSON(&body)

	user := models.User{
		Name:    body.Name,
		Email:   body.Email,
		FamiyId: body.FamilyId,
		Role:    body.Role,
	}

	query := db.NewInsert().Model(&user)
	sqlString := query.String()
	_, err := query.Exec(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"debug": gin.H{
				"sql": sqlString,
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
		"debug": gin.H{
			"sql": sqlString,
		},
	})
}
