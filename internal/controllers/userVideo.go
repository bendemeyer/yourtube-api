package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"yourtube/internal/models"
	"yourtube/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func getUserVideosQuery(userId int32) *bun.SelectQuery {
	db := repositories.GetDb()
	return db.NewSelect().
		Model((*models.Video)(nil)).
		ExcludeColumn("channel_id").
		ColumnExpr(`"uvv"."progress" as "progress"`).
		Relation("Channel").
		Join(`LEFT JOIN "family_allowed_channels" AS "fac" ON "channel"."id" = "fac"."channel_id" `).
		Join(`LEFT JOIN "family_allowed_videos" AS "fav" ON "video"."id" = "fav"."video_id"`).
		Join(`LEFT JOIN "family_blocked_videos" AS "fbv" ON "video"."id" = "fbv"."video_id"`).
		Join(`LEFT JOIN "user_allowed_channels" AS "uac" ON "channel"."id" = "uac"."channel_id"`).
		Join(`LEFT JOIN "user_blocked_channels" AS "ubc" ON "channel"."id" = "ubc"."channel_id"`).
		Join(`LEFT JOIN "user_allowed_videos" AS "uav" ON "video"."id" = "uav"."video_id"`).
		Join(`LEFT JOIN "user_blocked_videos" AS "ubv" ON "video"."id" = "ubv"."video_id"`).
		Join(`LEFT JOIN "families" AS "f" ON "fac"."family_id" = "f"."id" OR "fav"."family_id" = "f"."id" OR "fbv"."family_id" = "f"."id"`).
		Join(`LEFT JOIN "users" AS "u" ON "f"."id" = "u"."family_id" OR "uac"."user_id" = "u"."id" OR "ubc"."user_id" = "u"."id" OR "uav"."user_id" = "u"."id" OR "ubv"."user_id" = "u"."id"`).
		Join(`LEFT JOIN "user_video_views" AS "uvv" ON "video"."id" = "uvv"."video_id" AND "u"."id" = "uvv"."user_id"`).
		Where(`"u"."id" = ?`, userId).
		Where(`"fbv"."video_id" IS NULL`).
		Where(`"ubc"."channel_id" IS NULL`).
		Where(`"ubv"."video_id" IS NULL`)
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
	if ctx.Request.URL.Query().Get("history") == "1" {
		query = query.Where("progress > 0").OrderExpr("uvv.timestamp DESC")
	} else {
		query = handleQueryString(query, ctx.Request.URL.Query())
	}
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

	db := repositories.GetDb()
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
		OrderExpr(`"uvv"."timestamp" DESC`).
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
