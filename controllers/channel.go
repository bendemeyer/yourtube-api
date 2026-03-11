package controllers

import (
	"net/http"
	"strconv"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func GetChannel(ctx *gin.Context) {
	db := sqldb.GetDb()
	channel := new(models.Channel)
	err := db.NewSelect().Model(&channel).Where("id = ?", ctx.Param("channel_id")).Scan(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if channel == (&models.Channel{}) {
		ctx.JSON(http.StatusNotFound, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, channel)
}

func GetChannels(ctx *gin.Context) {
	db := sqldb.GetDb()
	var channels []models.Channel

	query := db.NewSelect().Model((*models.Channel)(nil))

	if handles, exists := ctx.GetQueryArray("handle"); exists {
		query = query.Where("channel.handle IN (?)", bun.In(handles))
	}

	if ids, exists := ctx.GetQueryArray("id"); exists {
		query = query.Where("channel.id IN (?)", bun.In(ids))
	}

	if token, exists := ctx.GetQuery("pageToken"); exists {
		query = query.Where("channel.id > ?", token)
	}

	query = query.OrderExpr("channel.id ASC")

	limit := 20
	if size, exists := ctx.GetQuery("size"); exists {
		if newLimit, err := strconv.Atoi(size); err != nil {
			limit = newLimit
		}
	}
	query = query.Limit(limit)

	sqlString := query.String()
	count, err := query.ScanAndCount(ctx, &channels)

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

	response := gin.H{
		"success":      true,
		"error":        false,
		"remainingResults": count,
		"channels":     channels,
		"debug": gin.H{
			"sql": sqlString,
		},
	}

	if count > len(channels) {
		response["nextPageToken"] = channels[len(channels) - 1].Id
	}

	ctx.JSON(http.StatusOK, response)
}

func PutChannel(ctx *gin.Context) {
	var channelModel models.Channel
	ctx.BindJSON(&channelModel)
	db := sqldb.GetDb()
	result, err := db.NewInsert().Model(&channelModel).On("CONFLICT (id) DO UPDATE").Exec(ctx)
	rows, _ := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"success":      true,
			"rowsAffected": rows,
		})
	}
}
