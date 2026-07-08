package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func getLatestVideoByChannel(channel models.Channel) (models.Video, error) {
	db := sqldb.GetDb()
	result := models.Video{}
	err := db.NewSelect().
		Model(&result).
		Where("channel_id = ?", channel.Id).
		OrderExpr("published DESC").
		Limit(1).
		Scan(context.Background())
	if err != nil {
		return models.Video{}, err
	}
	return result, nil
}

func upsertVideo(video models.Video) (sql.Result, error) {
	db := sqldb.GetDb()
	result, err := db.NewInsert().Model(video).On("CONFLICT (id) DO UPDATE").Exec(context.Background())
	return result, err
}

func handleQueryString(query *bun.SelectQuery, queryString url.Values) *bun.SelectQuery {
	orderColumn := "video.published"

	if queryString.Has("q") {
		query = query.ColumnExpr("ts_rank_cd(video.search_document, websearch_to_tsquery('english', ?)) AS search_rank", queryString.Get("q"))
		query = query.Where("video.search_document @@ websearch_to_tsquery('english', ?)", queryString.Get("q"))
		orderColumn = "search_rank"
	}

	basics := map[string]string{
		"shorterThan": "video.duration < ?",
		"longerThan":  "video.duration > ?",
	}
	for key, value := range basics {
		if queryString.Has(key) {
			param := queryString.Get(key)
			query = query.Where(value, param)
		}
	}

	timestamps := map[string]string{
		"before": "video.published < ?",
		"after":  "video.published > ?",
	}
	for key, value := range timestamps {
		if queryString.Has(key) {
			timestamp, err := time.Parse(time.RFC3339, queryString.Get(key))
			if err == nil {
				query = query.Where(value, timestamp)
			}
		}
	}

	arrays := map[string]string{
		"id":      "video.id IN (?)",
		"channel": "video.channel_id IN (?)",
	}
	for key, value := range arrays {
		if queryString.Has(key) {
			arr := queryString[key]
			query = query.Where(value, bun.In(arr))
		}
	}

	if queryString.Has("type") {
		if queryString.Get("type") == "video" {
			query = query.Where("video.is_short = ?", false)
		} else if queryString.Get("type") == "short" {
			query = query.Where("video.is_short = ?", true)
		}
	}

	orderExpr := fmt.Sprintf("%s DESC, video.id DESC", orderColumn)
	query = query.OrderExpr(orderExpr)

	return query
}

func paginate(query *bun.SelectQuery, page int, pageSize int) *bun.SelectQuery {
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize)
	return query
}

func GetVideo(ctx *gin.Context) {
	db := sqldb.GetDb()
	video := new(models.Video)
	err := db.NewSelect().Model(video).Where("video.id = ?", ctx.Param("video_id")).Relation("Channel").Scan(ctx)
	if err != nil {
		status := http.StatusInternalServerError
		if err == sql.ErrNoRows {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"video":   video,
	})
}

func GetVideos(ctx *gin.Context) {
	db := sqldb.GetDb()
	var videos []models.VideoResult

	pageSize := 20
	page, err := strconv.Atoi(ctx.Query(("page")))
	if err != nil {
		page = 1
	}

	query := db.NewSelect().Model((*models.Video)(nil)).ExcludeColumn("channel_id")
	fmt.Println(query.String())
	query = handleQueryString(query, ctx.Request.URL.Query())
	query = paginate(query, page, pageSize)
	query = query.Relation("Channel", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.ColumnExpr("channel.id AS channel__id").
			ColumnExpr("channel.handle AS channel__handle").
			ColumnExpr("channel.title AS channel__title").
			ColumnExpr("channel.thumbnails AS channel__thumbnails")
	})

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
		"success":          true,
		"error":            false,
		"remainingResults": count,
		"videos":           videos,
		"debug": gin.H{
			"sql": sqlString,
		},
	}

	if (count - pageSize) > pageSize {
		response["nextPage"] = page + 1
	}

	ctx.JSON(http.StatusOK, response)
}

func PutVideo(ctx *gin.Context) {
	var videoModel models.Video
	ctx.BindJSON(&videoModel)
	db := sqldb.GetDb()
	result, err := db.NewInsert().Model(&videoModel).On("CONFLICT (id) DO UPDATE").Exec(ctx)
	rows, _ := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"success":      true,
			"error":        false,
			"rowsAffected": rows,
		})
	}
}
