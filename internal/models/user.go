package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	Id              int32            `json:"id" bun:",pk,autoincrement"`
	Email           string           `json:"email" bun:",unique"`
	Name            string           `json:"name"`
	FamiyId         string           `json:"familyId" bun:"family_id"`
	Family          *Family          `json:"family" bun:"rel:belongs-to,join:family_id=id"`
	Role            FamilyMemberRole `json:"role"`
	AllowedChannels []*Channel       `json:"allowedChannels,omitempty" bun:"m2m:user_allowed_channels,join:User=Channel"`
	BlockedChannels []*Channel       `json:"blockedChannels,omitempty" bun:"m2m:user_blocked_channels,join:User=Channel"`
	AllowedVideos   []*Video         `json:"allowedVideos,omitempty" bun:"m2m:user_allowed_videos,join:User=Video"`
	BlockedVideos   []*Video         `json:"blockedVideos,omitempty" bun:"m2m:user_blocked_videos,join:User=Video"`
	History         []*UserVideoView `json:"history,omitempty" bun:"rel:has-many,join:id=user_id"`
}

type UserAllowedChannel struct {
	bun.BaseModel `bun:"table:user_allowed_channels"`

	UserId    int32    `json:"userId" bun:",pk"`
	User      *User    `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	ChannelId string   `json:"channelId" bun:",pk"`
	Channel   *Channel `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
}

type UserBlockedChannel struct {
	bun.BaseModel `bun:"table:user_blocked_channels"`

	UserId    int32    `json:"userId" bun:",pk"`
	User      *User    `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	ChannelId string   `json:"channelId" bun:",pk"`
	Channel   *Channel `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
}

type UserAllowedVideo struct {
	bun.BaseModel `bun:"table:user_allowed_videos"`

	UserId  int32  `json:"userId" bun:",pk"`
	User    *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	VideoId string `json:"videoId" bun:",pk"`
	Video   *Video `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}

type UserBlockedVideo struct {
	bun.BaseModel `bun:"table:user_blocked_videos"`

	UserId  int32  `json:"userId" bun:",pk"`
	User    *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	VideoId string `json:"videoId" bun:",pk"`
	Video   *Video `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}

type UserVideoView struct {
	bun.BaseModel `bun:"table:user_video_views"`

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
