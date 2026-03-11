package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/net/html"
)

func getUserVideosQuery(userId int32) *bun.SelectQuery {
	db := sqldb.GetDb()
	return db.NewSelect().
		Model((*models.Video)(nil)).
		ExcludeColumn("channel_id").
		ColumnExpr(`"video_view"."progress" as "progress"`).
		Relation("Channel").
		Join(`LEFT JOIN "allowed_videos" AS "allowed_video" ON "video"."id" = "allowed_video"."video_id"`).
		Join(`LEFT JOIN "allowed_channels" AS "allowed_channel" ON "channel"."id" = "allowed_channel"."channel_id"`).
		Join(`LEFT JOIN "users" AS "user" ON "allowed_video"."user_id" = "user"."id" OR "allowed_channel"."user_id" = "user"."id"`).
		Join(`LEFT JOIN "blocked_videos" AS "blocked_video" ON "video"."id" = "blocked_video"."video_id" AND "user"."id" = "blocked_video"."user_id"`).
		Join(`LEFT JOIN "video_views" AS "video_view" ON "video"."id" = "video_view"."video_id" AND "user"."id" = "video_view"."user_id"`).
		Where(`"blocked_video"."video_id" IS NULL`).
		Where(`"user"."id" = ?`, userId)
}

func GetVideoPlayer(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.Data(
			http.StatusBadRequest,
			"text/html",
			[]byte(fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id"))),
		)
		return
	}
	count, err := getUserVideosQuery(int32(userId)).Where("video.id = ?", ctx.Param("video_id")).Count(ctx)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}
	if count == 0 {
		ctx.Data(http.StatusNotFound, "text/html", nil)
		return
	}
	embedUrl := fmt.Sprintf("https://www.youtube.com/embed/%s?enablejsapi=1", ctx.Param("video_id"))
	response, err := http.Get(embedUrl)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			ctx.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
			return
		}
		ctx.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
		return
	}

	htmlDoc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}

	baseTag := &html.Node{
		Type: html.ElementNode,
		Data: "base",
		Attr: []html.Attribute{
			{Key: "href", Val: "https://www.youtube.com/embed/"},
		},
	}

	htmlDoc.Find("head").First().PrependNodes(baseTag)

	styleTag := &html.Node{
		Type: html.ElementNode,
		Data: "style",
		Attr: []html.Attribute{
			{Key: "type", Val: "text/css"},
		},
	}

	styleContent := &html.Node{
		Type: html.TextNode,
		Data: ".ytp-pause-overlay, .ytp-chrome-top, .ytp-youtube-button { display: none !important; }",
	}

	styleTag.AppendChild(styleContent)
	htmlDoc.Find("head").First().AppendNodes(styleTag)

	result, err := htmlDoc.Html()
	if err != nil {
		ctx.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}
	ctx.Data(http.StatusOK, "text/html", []byte(result))
}

func GetUserVideos(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Bad Request: User ID parameter %s is not an integer", ctx.Param("user_id")),
		},
		)
		return
	}

	var videos []models.UserVideoResult
	query := getUserVideosQuery(int32(userId)).Limit(20)
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

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"count":   count,
		"videos":  videos,
		"debug":   sqlString,
	})
}
