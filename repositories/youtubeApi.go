package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"yourtube/models"

	"github.com/sosodev/duration"
)

type YtVideosResponse struct {
	Items []YtVideo `json:"items"`
}

type YtVideo struct {
	Id             string                `json:"id"`
	Kind           string                `json:"kind"`
	Snippet        YtVideoSnippet        `json:"snippet"`
	ContentDetails YtVideoContentDetails `json:"contentDetails"`
}

type YtVideoSnippet struct {
	ChannelId   string                        `json:"channelId"`
	Title       string                        `json:"title"`
	Description string                        `json:"description"`
	PublishedAt *time.Time                    `json:"publishedAt"`
	Thumbnails  map[string]YtThumbnailDetails `json:"thumbnails"`
	CategoryId  string                        `json:"categoryId"`
	Tags        []string                      `json:"tags"`
}

type YtVideoContentDetails struct {
	Duration string `json:"duration"`
}

type YtPlaylistItemsResponse struct {
	Items         []YtPlaylistItem `json:"items"`
	NextPageToken string           `json:"nextPageToken"`
}

type YtPlaylistItem struct {
	Id      string                `json:"id"`
	Snippet YtPlaylistItemSnippet `json:"snippet"`
}

type YtPlaylistItemSnippet struct {
	ResourceId  YtPlaylistItemResourceId `json:"resourceId"`
	PublishedAt *time.Time               `json:"publishedAt"`
}

type YtPlaylistItemResourceId struct {
	Kind    string `json:"kind"`
	VideoId string `json:"videoId"`
}

type YtChannelsResponse struct {
	Items []YtChannel `json:"items"`
}

type YtChannel struct {
	Id      string           `json:"id"`
	Snippet YtChannelSnippet `json:"snippet,omitempty"`
}

type YtChannelSnippet struct {
	CustomUrl   string                        `json:"customUrl,omitempty"`
	Title       string                        `json:"title,omitempty"`
	Description string                        `json:"description,omitempty"`
	Thumbnails  map[string]YtThumbnailDetails `json:"thumbnails,omitempty"`
}

type YtThumbnailDetails struct {
	Url    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

func buildRequestUrl(endpoint string, params url.Values) string {
	key := os.Getenv("YOUTUBE_DATA_API_KEY")
	queryString := params.Encode()
	return fmt.Sprintf("https://www.googleapis.com/youtube/v3/%s?key=%s&%s", endpoint, key, queryString)

}

func getPlaylistItems(playlistId string, pageToken string) (YtPlaylistItemsResponse, error) {
	urlParams := url.Values{
		"playlistId":  {playlistId},
		"part":        {"snippet"},
		"max_results": {"50"},
	}
	if pageToken != "" {
		urlParams.Set("pageToken", pageToken)
	}
	response, err := http.Get(buildRequestUrl("playlistItems", urlParams))
	if err != nil || response.StatusCode != http.StatusOK {
		return YtPlaylistItemsResponse{}, errors.New("Unable to fetch videos for playlist")
	}
	defer response.Body.Close()

	var body YtPlaylistItemsResponse
	json_err := json.NewDecoder(response.Body).Decode(&body)
	if json_err != nil {
		return YtPlaylistItemsResponse{}, json_err
	}

	return body, nil
}

func getVideos(videoIds []string, shorts bool) ([]models.Video, error) {
	url := buildRequestUrl("videos", url.Values{
		"part": {"snippet,contentDetails"},
		"id":   {strings.Join(videoIds, ",")},
	})
	response, err := http.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		return []models.Video{}, errors.New("Unable to get channel details from YouTube Data API")
	}
	defer response.Body.Close()

	var body YtVideosResponse
	json_err := json.NewDecoder(response.Body).Decode(&body)
	if json_err != nil {
		return []models.Video{}, json_err
	}

	results := []models.Video{}
	for _, video := range body.Items {
		categoryIntId, _ := strconv.Atoi(video.Snippet.CategoryId)
		vidDuration, _ := duration.Parse(video.ContentDetails.Duration)
		results = append(results, models.Video{
			Id:          video.Id,
			ChannelId:   video.Snippet.ChannelId,
			CategoryId:  int8(categoryIntId),
			Title:       video.Snippet.Title,
			Description: video.Snippet.Description,
			Published:   video.Snippet.PublishedAt,
			Duration:    int32(vidDuration.Seconds),
			IsShort:     shorts,
			Tags:        video.Snippet.Tags,
			Thumbnails: []string{
				video.Snippet.Thumbnails["default"].Url,
				video.Snippet.Thumbnails["medium"].Url,
				video.Snippet.Thumbnails["high"].Url,
				video.Snippet.Thumbnails["standard"].Url,
				video.Snippet.Thumbnails["maxres"].Url,
			},
		})
	}
	return results, nil
}

func generateShortIdsByChannel(channel models.Channel) iter.Seq[string] {
	return func(yield func(string) bool) {
		complete := false
		buffer := []string{}
		nextPageToken := ""

		for complete == false || len(buffer) > 0 {
			if len(buffer) == 0 {
				response, _ := getPlaylistItems(channel.GetShortsPlaylist(), nextPageToken)
				for _, value := range response.Items {
					buffer = append(buffer, value.Snippet.ResourceId.VideoId)
				}
				if response.NextPageToken != "" {
					nextPageToken = response.NextPageToken
				} else {
					complete = true
				}
			}
			next := buffer[0]
			buffer = buffer[1:]
			if !yield(next) {
				return
			}
		}
	}
}

func GenerateVideosByChannel(channel models.Channel, after time.Time) iter.Seq[models.Video] {
	return func(yield func(models.Video) bool) {
		shorts := generateShortIdsByChannel(channel)
		nextShort, stopShorts := iter.Pull(shorts)
		defer stopShorts()
		latestShortId, _ := nextShort()

		complete := false
		buffer := []models.Video{}
		nextPageToken := ""

		for complete == false || len(buffer) > 0 {
			if len(buffer) == 0 {
				videoIds := []string{}
				response, _ := getPlaylistItems(channel.GetShortsPlaylist(), nextPageToken)
				if response.NextPageToken != "" {
					nextPageToken = response.NextPageToken
				} else {
					complete = true
				}
				for _, video := range response.Items {
					if video.Snippet.ResourceId.VideoId == latestShortId {
						latestShortId, _ = nextShort()
						continue
					}
					if !video.Snippet.PublishedAt.After(after) || video.Snippet.ResourceId.Kind != "youtube#video" {
						continue
					}
					videoIds = append(videoIds, video.Snippet.ResourceId.VideoId)
				}
				videos, _ := getVideos(videoIds, false)
				for _, video := range videos {
					buffer = append(buffer, video)
				}
			}
			next := buffer[0]
			buffer = buffer[1:]
			if !yield(next) {
				return
			}
		}
	}
}

func GetChannel(channelId string) (models.Channel, error) {
	url := buildRequestUrl("channels", url.Values{
		"part": {"snippet"},
		"id":   {channelId},
	})
	response, err := http.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		return models.Channel{}, errors.New("Unable to get channel details from YouTube Data API")
	}
	defer response.Body.Close()

	var body YtChannelsResponse
	json_err := json.NewDecoder(response.Body).Decode(&body)
	if json_err != nil {
		return models.Channel{}, json_err
	}

	ytChannel := body.Items[0]
	return models.Channel{
		Id:          ytChannel.Id,
		Handle:      ytChannel.Snippet.CustomUrl,
		Title:       ytChannel.Snippet.Title,
		Description: ytChannel.Snippet.Description,
		Thumbnails: []string{
			ytChannel.Snippet.Thumbnails["default"].Url,
			ytChannel.Snippet.Thumbnails["medium"].Url,
			ytChannel.Snippet.Thumbnails["high"].Url,
		},
	}, nil
}

func getChannelIdFromHandle(handle string) (string, error) {
	url := buildRequestUrl("channels", url.Values{
		"part":      {"id"},
		"forHandle": {handle},
	})
	response, err := http.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		return "", errors.New("Unable to get channel details from YouTube Data API")
	}
	defer response.Body.Close()

	var body YtChannelsResponse
	json_err := json.NewDecoder(response.Body).Decode(&body)
	if json_err != nil {
		return "", json_err
	}

	return body.Items[0].Id, nil
}

func GetChannelFromHandle(handle string) (models.Channel, error) {
	channelId, err := getChannelIdFromHandle(handle)
	if err != nil {
		return models.Channel{}, err
	}
	return GetChannel(channelId)
}
