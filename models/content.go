package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Channel struct {
	bun.BaseModel `bun:"table:channels"`

	Id          string      `json:"id" bun:",pk,unique"`
	Handle      string      `json:"handle" bun:",unique,notnull"`
	Title       string      `json:"title" bun:",notnull"`
	Description string      `json:"description"`
	Thumbnails  []string    `json:"thumbnails" bun:",array"`
	Videos      []*Video    `json:"videos,omitempty" bun:"rel:has-many,join:id=channel_id"`
	Playlists   []*Playlist `json:"playlists,omitempty" bun:"rel:has-many,join:id=channel_id"`
}

func (c Channel) GetUploadsPlaylist() string {
	return "UU" + c.Id[2:]
}

func (c Channel) GetShortsPlaylist() string {
	return "UUSH" + c.Id[2:]
}

type Video struct {
	bun.BaseModel `bun:"table:videos"`

	Id          string      `json:"id" bun:",pk,unique"`
	ChannelId   string      `json:"channelId,omitempty" bun:",notnull"`
	Channel     *Channel    `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
	CategoryId  int8        `json:"categoryId" bun:",notnull"`
	Category    *Category   `json:"category,omitempty" bun:"rel:has-one,join:category_id=id"`
	Title       string      `json:"title" bun:",notnull"`
	Description string      `json:"description"`
	Duration    int32       `json:"duration" bun:",notnull"`
	Thumbnails  []string    `json:"thumbnails" bun:",array"`
	IsShort     bool        `json:"isShort" bun:",default:false"`
	Published   *time.Time  `json:"published"`
	InfoLinks   []string    `json:"infoLinks" bun:",array"`
	Tags        []string    `json:"tags" bun:",array"`
	Playlists   []*Playlist `json:"playlists,omitempty" bun:"m2m:playlist_videos,join:Video=Playlist"`
}

type Playlist struct {
	bun.BaseModel `bun:"table:playlists"`

	Id          string   `json:"id" bun:",pk,unique"`
	ChannelId   string   `json:"channelId" bun:",notnull"`
	Channel     *Channel `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
	Title       string   `json:"title" bun:",notnull"`
	Description string   `json:"description"`
	Thumbnails  []string `json:"thumbnails" bun:",array"`
	Videos      []*Video `json:"videos,omitempty" bun:"m2m:playlist_videos,join:Playlist=Video"`
}

type PlaylistVideo struct {
	bun.BaseModel `bun:"table:playlist_videos"`

	VideoId    string    `json:"videoId" bun:",pk"`
	Video      *Video    `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
	PlaylistId string    `json:"playlistId" bun:",pk"`
	Playlist   *Playlist `json:"playlist,omitempty" bun:"rel:belongs-to,join:playlist_id=id"`
}

type Category struct {
	bun.BaseModel `bun:"table:categories"`

	Id     int8     `json:"id" bun:",pk,unique"`
	Title  string   `json:"title" bun:",notnull"`
	Videos []*Video `json:"videos,omitempty" bun:"rel:has-many,join:id=category_id"`
}

type VideoResult struct {
	Video

	SearchRank float32 `json:"-"`
}
