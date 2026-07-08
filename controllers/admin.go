package controllers

import (
	"net/http"
	"yourtube/models"
	"yourtube/repositories"

	"github.com/gin-gonic/gin"
)

func updateChannelVideos(channel models.Channel) error {
	latestVid, err := getLatestVideoByChannel(channel)
	if err != nil {
		return err
	}
	videos := repositories.GenerateVideosByChannel(channel, *latestVid.Published)
	for video := range videos {
		upsertVideo(video)
	}
	return nil
}

func updateAllChannelVideos() {
	channels := generateAllChannels()
	for channel := range channels {
		updateChannelVideos(channel)
	}
}

func DoAdminAction(ctx *gin.Context) {
	type AdminAction string
	const (
		UpdateAllChannelVideos AdminAction = "update_all_channel_videos"
	)
	type AdminRequestBody struct {
		Action AdminAction `json:"action"`
	}
	var body AdminRequestBody
	err := ctx.ShouldBindJSON(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}

	if body.Action == UpdateAllChannelVideos {
		updateAllChannelVideos()
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}

}
