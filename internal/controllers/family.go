package controllers

import (
	"net/http"
	"yourtube/internal/models"
	"yourtube/internal/repositories"

	"github.com/gin-gonic/gin"
)

func AddFamily(ctx *gin.Context) {
	var family models.Family
	db := repositories.GetDb()
	err := db.NewInsert().Model(&family).Scan(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"family":  family,
	})
}
