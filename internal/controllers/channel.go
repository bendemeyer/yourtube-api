package controllers

import (
	"context"
	"database/sql"
	"errors"
	"iter"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"
	"yourtube/internal/models"
	"yourtube/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

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
	db := repositories.GetDb()
	exists, exists_err := channelExists(channel.Id)
	result, err := db.NewInsert().Model(&channel).On("CONFLICT (id) DO UPDATE").Exec(context.Background())
	if !exists && exists_err == nil && err == nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Safely recovered from goroutine panic: %v\n", r)
					log.Print(string(debug.Stack()))
				}
			}()
			bootstrapChannelVideos(channel)
		}()
	}
	return result, err
}

func channelExists(channelId string) (bool, error) {
	db := repositories.GetDb()
	channel := models.Channel{}
	err := db.NewSelect().Model(&channel).Where("id = ?", channelId).Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func getChannel(channelId string) (models.Channel, error) {
	db := repositories.GetDb()
	channel := models.Channel{}
	err := db.NewSelect().Model(&channel).Where("id = ?", channelId).Scan(context.Background())
	return channel, err
}

func getChannelByHandle(handle string) (models.Channel, error) {
	db := repositories.GetDb()
	channel := models.Channel{}
	err := db.NewSelect().Model(&channel).Where("handle = ?", handle).Scan(context.Background())
	return channel, err
}

func getChannels(offset int, limit int) ([]models.Channel, error) {
	db := repositories.GetDb()
	result := []models.Channel{}
	err := db.NewSelect().
		Model(&result).
		Offset(offset).
		Limit(limit).
		Scan(context.Background())
	return result, err
}

func generateAllChannels() iter.Seq[models.Channel] {
	return func(yield func(models.Channel) bool) {
		complete := false
		buffer := []models.Channel{}
		offset := 0
		batchSize := 1000

		for complete == false || len(buffer) > 0 {
			if len(buffer) == 0 {
				channels, err := getChannels(offset, batchSize)
				if err != nil {
					continue
				}
				if len(channels) < batchSize {
					complete = true
				}
				buffer = channels
				offset += batchSize
			}
			next := buffer[0]
			buffer = buffer[1:]
			if !yield(next) {
				return
			}
		}
	}
}

func deleteChannel(channel models.Channel) (sql.Result, error) {
	db := repositories.GetDb()
	var channelVideos []models.Video
	err := db.NewSelect().Model(&channelVideos).Where("channel_id = ?", channel.Id).Scan(context.Background())
	if err != nil {
		return *new(sql.Result), err
	}
	for _, video := range channelVideos {
		result, err := deleteVideo(video)
		if err != nil {
			return result, err
		}
	}
	var channelPlaylists []models.Playlist
	err = db.NewSelect().Model(&channelPlaylists).Where("channel_id = ?", channel.Id).Scan(context.Background())
	if err != nil {
		return *new(sql.Result), err
	}
	for _, playlist := range channelPlaylists {
		result, err := deletePlaylist(playlist)
		if err != nil {
			return result, err
		}
	}
	dependencies := []string{
		db.Table(reflect.TypeOf(models.FamilyAllowedChannel{})).Name,
		db.Table(reflect.TypeOf(models.UserAllowedChannel{})).Name,
		db.Table(reflect.TypeOf(models.UserBlockedChannel{})).Name,
	}
	for _, dependency := range dependencies {
		result, err := db.NewDelete().Table(dependency).Where("channel_id = ?", channel.Id).Exec(context.Background())
		if err != nil {
			return result, err
		}
	}
	return db.NewDelete().Model(&models.Channel{}).Where("id = ?", channel.Id).Exec(context.Background())
}

func deletePlaylist(playlist models.Playlist) (sql.Result, error) {
	db := repositories.GetDb()
	dependencies := []string{
		db.Table(reflect.TypeOf(models.PlaylistVideo{})).Name,
	}
	for _, dependency := range dependencies {
		result, err := db.NewDelete().Table(dependency).Where("playlist_id = ?", playlist.Id).Exec(context.Background())
		if err != nil {
			return result, err
		}
	}
	return db.NewDelete().Model(&playlist).Where("id = ?", playlist.Id).Exec(context.Background())
}

func GetChannel(ctx *gin.Context) {
	channel, err := getChannel(ctx.Param("channel_id"))
	if errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, gin.H{})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, channel)
}

func GetChannels(ctx *gin.Context) {
	db := repositories.GetDb()
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

func PutChannel(ctx *gin.Context) {
	var channelModel models.Channel
	ctx.ShouldBindJSON(&channelModel)
	result, err := upsertChannel(channelModel)
	rows, _ := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"success":      true,
			"rowsAffected": rows,
		})
	}
}

func AddChannel(ctx *gin.Context) {
	type AddChannelPostBody struct {
		Handle string `json:"handle"`
	}
	var body AddChannelPostBody
	body_err := ctx.ShouldBindJSON(&body)
	if body_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   body_err.Error(),
		})
		return
	}

	_, err := getChannelByHandle(body.Handle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.New("Cannot add this channel, it already exists"),
		})
		return
	}

	channel, err := bootstrapChannel(body.Handle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"channel": channel,
	})
}

func DeleteChannel(ctx *gin.Context) {
	channel, err := getChannel(ctx.Param("channel_id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	_, err = deleteChannel(channel)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
