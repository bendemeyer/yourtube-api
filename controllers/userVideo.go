package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func getUserVideosQuery(userId int32) *bun.SelectQuery {
	db := sqldb.GetDb()
	return db.NewSelect().
		Model((*models.Video)(nil)).
		ExcludeColumn("channel_id").
		ColumnExpr(`"video_view"."progress" as "progress"`).
		Relation("Channel").
		Join(`LEFT JION "family_allowed_videos" AS "family_allowed_video" ON "video"."id" = "family_allowed_video"."video_id"`).
		Join(`LEFT JOIN "families" AS "family" ON "family_allowed_video"."family_id" = "family"."id`).
		Join(`LEFT JOIN "users" AS "user" ON "family"."id" = "user"."family_id"`).
		Join(`LEFT JOIN "family_allowed_channels" AS "family_allowed_channel" ON "channel"."id" = "family_allowed_channel"."channel_id"`).
		Join(`LEFT JOIN "family_blocked_videos" AS "family_blocked_video" ON "video"."id" = "family_blocked_video"."video_id" AND "family"."id" = "family_blocked_video"."family_id"`).
		Join(`LEFT JOIN "user_allowed_videos" AS "user_allowed_video" ON "video"."id" = "user_allowed_video"."video_id"`).
		Join(`LEFT JOIN "user_allowed_channels" AS "user_allowed_channel" ON "channel"."id" = "user_allowed_channel"."channel_id"`).
		Join(`LEFT JOIN "user_blocked_videos" AS "user_blocked_video" ON "video"."id" = "user_blocked_video"."video_id" AND "user"."id" = "user_blocked_video"."user_id"`).
		Join(`LEFT JOIN "user_blocked_channels" AS "user_blocked_channel" ON "channel"."id" = "user_blocked_channel"."channel_id" AND "user"."id" = "user_blocked_channel"."user_id"`).
		Join(`LEFT JOIN "video_views" AS "video_view" ON "video"."id" = "video_view"."video_id" AND "user"."id" = "video_view"."user_id"`).
		Where(`"family_blocked_video"."video_id" IS NULL`).
		Where(`"user_blocked_video"."video_id" IS NULL`).
		Where(`"user_blocked_channel"."channel_id" IS NULL`).
		Where(`"user"."id" = ?`, userId)
}

func GetUserVideos(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	pageSize := 20
	page, err := strconv.Atoi(ctx.Query(("page")))
	if err != nil {
		page = 1
	}

	var videos []models.UserVideoResult
	query := getUserVideosQuery(int32(userId))
	query = handleQueryString(query, ctx.Request.URL.Query())
	query = paginate(query, page, pageSize)
	sqlString := query.String()
	count, err := query.ScanAndCount(ctx, &videos)
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
		"success": true,
		"error":   false,
		"count":   count,
		"videos":  videos,
		"debug": gin.H{
			"sql": sqlString,
		},
	}

	if (count - pageSize) > pageSize {
		response["nextPage"] = page + 1
	}

	ctx.JSON(http.StatusOK, response)
}

func GetUserVideo(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	video := new(models.UserVideoResult)
	query := getUserVideosQuery(int32(userId)).Where("video.id = ?", ctx.Param("video_id"))
	sqlString := query.String()
	err := query.Scan(ctx)
	if err != nil {
		status := http.StatusInternalServerError
		if err == sql.ErrNoRows {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
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
		"error":   false,
		"video":   video,
		"debug": gin.H{
			"sql": sqlString,
		},
	})
}

func UpdateProgress(ctx *gin.Context) {
	var errors []string
	videoId := ctx.Param(("video_id"))

	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		errors = append(errors, fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")))
	}

	type VideoViewRequestBody struct {
		Progress int32 `json:"progress"`
	}
	var request VideoViewRequestBody
	body_err := ctx.ShouldBindJSON(&request)
	if body_err != nil {
		errors = append(errors, body_err.Error())
	}

	if len(errors) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errors":  errors,
		})
		return
	}

	db := sqldb.GetDb()
	view := &models.UserVideoView{
		UserId:    int32(userId),
		VideoId:   videoId,
		Progress:  request.Progress,
		Timestamp: time.Now(),
	}
	_, err := db.NewInsert().Model(view).
		On("CONFLICT (user_id, video_id) DO UPDATE").Exec(ctx)

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

func GetViewedVideos(ctx *gin.Context) {
	userId, id_err := strconv.Atoi(ctx.Param("user_id"))
	if id_err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		})
		return
	}

	var videos []models.UserVideoResult
	query := getUserVideosQuery(int32(userId)).
		Where("progress > 0").
		OrderExpr(`"video_view"."timestamp" DESC`).
		Limit(20)

	sqlString := query.String()
	sql_err := query.Scan(ctx, &videos)

	if sql_err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   sql_err.Error(),
			"debug": gin.H{
				"sql": sqlString,
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"videos":  videos,
		"debug": gin.H{
			"sql": sqlString,
		},
	})

}
