package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"yourtube/internal/models"
	"yourtube/internal/repositories"

	"github.com/gin-gonic/gin"
)

func AllowUserChannel(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	type AddUserChannelRequestBody struct {
		Handle string `json:"handle"`
	}
	var body AddUserChannelRequestBody
	body_err := ctx.ShouldBindJSON(&body)
	if body_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   body_err.Error(),
		})
		return
	}

	var channel models.Channel
	existing, exists_err := getChannelByHandle(body.Handle)
	if errors.Is(exists_err, sql.ErrNoRows) {
		c, err := bootstrapChannel(body.Handle)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
		}
		channel = c
	} else {
		channel = existing
	}

	db := repositories.GetDb()
	userChannel := &models.UserAllowedChannel{
		UserId:    int32(userId),
		ChannelId: channel.Id,
	}
	query := db.NewInsert().Model(userChannel)
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

func DeleteAllowedUserChannel(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	channelId := ctx.Param("channel_id")
	db := repositories.GetDb()

	userChannel := &models.UserAllowedChannel{
		UserId:    int32(userId),
		ChannelId: channelId,
	}
	query := db.NewDelete().Model(userChannel)
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
