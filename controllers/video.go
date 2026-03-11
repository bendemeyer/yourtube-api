package controllers

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"yourtube/models"
	"yourtube/sqldb"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func generatePageToken(video models.VideoResult) (string, error) {
	var bits uint64

	// If the results are ranked, the ranking field determines sort order
	if video.SearchRank > 0 {
		bits = uint64(math.Float32bits(video.SearchRank))
		// If not ranked, sort order is determined by published time
	} else {
		bits = uint64(video.Published.Unix())
	}
	slice := make([]byte, 8)
	// trim leading zeros to save space
	binary.BigEndian.PutUint64(slice, bits)
	for slice[0] == 0 {
		slice = slice[1:]
	}

	decodedId, err := base64.RawURLEncoding.DecodeString(video.Id)
	if err != nil {
		return "", err
	}

	// concat timestamp & vidId arrays
	combined := append(decodedId, slice...)
	// encode the whole thing in base64 for usage in URLs
	token := base64.RawURLEncoding.EncodeToString(combined)
	return token, nil
}

func parsePageToken(token string) (uint64, string, error) {
	decodedToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, "", err
	}

	// first 8 bytes are the video ID
	idSlice := decodedToken[:8]
	vidId := base64.RawURLEncoding.EncodeToString(idSlice)

	// remaining bytes are the sort value & direction
	slice := decodedToken[8:]
	// pad leading zeros until we have 8 bytes for a uint64
	for len(slice) < 8 {
		slice = append([]byte{0}, slice...)
	}
	// convert to uint64
	bits := binary.BigEndian.Uint64(slice)
	return bits, vidId, nil
}

func handleQueryString(query *bun.SelectQuery, queryString url.Values) (*bun.SelectQuery) {
	orderColumn := "video.published"
	bitsProcessor := func(bits uint64) interface{} {
		return time.Unix(int64(bits), 0)
	}

	if queryString.Has("q") {
		query = query.ColumnExpr("ts_rank_cd(video.search_document, websearch_to_tsquery('english', ?)) AS search_rank", queryString.Get("q"))
		query = query.Where("video.search_document @@ websearch_to_tsquery('english', ?)", queryString.Get("q"))
		orderColumn = "search_rank"
		bitsProcessor = func(bits uint64) interface{} {
			return math.Float32frombits(uint32(bits))
		}
	}

	if queryString.Has("pageToken") {
		bits, vidId, err := parsePageToken(queryString.Get("pageToken"))
		sortValue := bitsProcessor(bits)
		if err == nil {
			expr := fmt.Sprintf("(%s, video.id) < (?, ?)", orderColumn)
			query = query.Where(expr, sortValue, vidId)
		}
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

	limit := 20
	if queryString.Has("size") {
		size, err := strconv.Atoi(queryString.Get("size"))
		if err == nil {
			if size > 100 {
				size = 100
			}
			limit = size
		}
	}
	query = query.Limit(limit)

	orderExpr := fmt.Sprintf("%s DESC, video.id DESC", orderColumn)
	query = query.OrderExpr(orderExpr)

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

	query := db.NewSelect().Model((*models.Video)(nil)).ExcludeColumn("channel_id")
	fmt.Println(query.String())
	query = handleQueryString(query, ctx.Request.URL.Query())
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
		"success":      true,
		"error":        false,
		"remainingResults": count,
		"videos":       videos,
		"debug": gin.H{
			"sql": sqlString,
		},
	}
	if count > len(videos) {
		nextPageToken, err := generatePageToken(videos[len(videos)-1])
		if err == nil {
			response["nextPageToken"] = nextPageToken
		} else {
			log.Println("Error generating page token: ", err)
		}
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
