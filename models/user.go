package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	Id            int32        `json:"id" bun:",pk,autoincrement"`
	Email         string       `json:"email" bun:",unique"`
	ParentId      int32        `json:"parentId"`
	Children      []*User      `json:"children,omitempty" bun:"rel:has-many,join:id=parent_id"`
	Channels      []*Channel   `json:"channels,omitempty" bun:"m2m:allowed_channels,join:User=Channel"`
	AllowedVideos []*Video     `json:"allowedVideos,omitempty" bun:"m2m:allowed_videos,join:User=Video"`
	BlockedVideos []*Video     `json:"blockedVideos,omitempty" bun:"m2m:blocked_videos,join:User=Video"`
	History       []*VideoView `json:"history,omitempty" bun:"rel:has-many,join:id=user_id"`
}

type AllowedChannel struct {
	bun.BaseModel `bun:"table:allowed_channels"`

	UserId    int32    `json:"userId" bun:",pk"`
	User      *User    `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	ChannelId string   `json:"channelId" bun:",pk"`
	Channel   *Channel `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
}

type AllowedVideo struct {
	bun.BaseModel `bun:"table:allowed_videos"`

	UserId  int32  `json:"userId" bun:",pk"`
	User    *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	VideoId string `json:"videoId" bun:",pk"`
	Video   *Video `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}

type BlockedVideo struct {
	bun.BaseModel `bun:"table:blocked_videos"`

	UserId  int32  `json:"userId" bun:",pk"`
	User    *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	VideoId string `json:"videoId" bun:",pk"`
	Video   *Video `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}

type VideoView struct {
	bun.BaseModel `bun:"table:video_views"`

	UserId    int32     `json:"userId" bun:",pk"`
	User      *User     `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	VideoId   string    `json:"videoId" bun:",pk"`
	Video     *Video    `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
	Timestamp time.Time `json:"timestamp"`
	Progress  int32     `json:"progress"`
}

type UserVideoResult struct {
	VideoResult

	Progress int32 `json:"progress"`
}
