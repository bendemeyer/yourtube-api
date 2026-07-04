package controllers

import (
	"net/http"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
)

func AddUser(ctx *gin.Context) {
	db := sqldb.GetDb()

	type AddUserRequestBody struct {
		Email string `json:"email"`
	}
	var body AddUserRequestBody
	ctx.ShouldBindJSON(body)

	user := models.User{
		Email: body.Email,
	}

	query := db.NewInsert().Model(user)
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
		"debug": gin.H{
			"sql": sqlString,
		},
	})
}
