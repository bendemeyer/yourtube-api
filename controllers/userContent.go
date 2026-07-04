package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
)

func AddUserChannel(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	channelId := ctx.Param("channel_id")
	db := sqldb.GetDb()

	userChannel := &models.AllowedChannel{
		UserId:    int32(userId),
		ChannelId: channelId,
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

func DeleteUserChannel(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	channelId := ctx.Param("channel_id")
	db := sqldb.GetDb()

	userChannel := &models.AllowedChannel{
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
