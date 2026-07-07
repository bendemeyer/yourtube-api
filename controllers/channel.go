package controllers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"yourtube/models"
	"yourtube/repositories"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func getChannel(channelId string) ([]models.Channel, error) {
	db := sqldb.GetDb()
	channels := []models.Channel{}
	err := db.NewSelect().Model(models.Channel{}).Where("id = ?", channelId).Scan(context.Background(), &channels)
	return channels, err
}

func getChannelByHandle(handle string) ([]models.Channel, error) {
	db := sqldb.GetDb()
	channels := []models.Channel{}
	err := db.NewSelect().Model(models.Channel{}).Where("handle = ?", handle).Scan(context.Background(), &channels)
	return channels, err
}

func GetChannel(ctx *gin.Context) {
	channels, err := getChannel(ctx.Param("channel_id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if len(channels) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, channels[0])
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
		"success":          true,
		"error":            false,
		"remainingResults": count,
		"channels":         channels,
		"debug": gin.H{
			"sql": sqlString,
		},
	}

	if count > len(channels) {
		response["nextPageToken"] = channels[len(channels)-1].Id
	}

	ctx.JSON(http.StatusOK, response)
}

func bootstrapChannel(handle string) (models.Channel, error) {
	channel, err := repositories.GetChannelFromHandle(handle)
	if err != nil {
		return models.Channel{}, err
	}
	_, err = upsertChannel(channel)
	return channel, err
}

func bootstrapChannelVideos(channel models.Channel) {
	videoGenerator := repositories.GenerateVideosByChannel(channel, time.Time{})
	for video := range videoGenerator {
		upsertVideo(video)
	}
}

func upsertChannel(channel models.Channel) (sql.Result, error) {
	db := sqldb.GetDb()
	exists, exists_err := getChannel(channel.Id)
	result, err := db.NewInsert().Model(channel).On("CONFLICT (id) DO UPDATE").Exec(context.Background())
	if len(exists) == 0 && err == nil && exists_err == nil {
		go bootstrapChannelVideos(channel)
	}
	return result, err
}

func PutChannel(ctx *gin.Context) {
	var channelModel models.Channel
	ctx.BindJSON(&channelModel)
	result, err := upsertChannel(channelModel)
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
