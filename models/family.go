package models

import (
	"github.com/uptrace/bun"
)

type FamilyMemberRole string

const (
	Adult FamilyMemberRole = "adult"
	Child FamilyMemberRole = "child"
)

type Family struct {
	bun.BaseModel `bun:"table:families"`

	Id              int32           `json:"id" bun:",pk,autoincrement"`
	Members         []*FamilyMember `json:"members,omitempty" bun:"m2m:family_members,join:Family=User"`
	AllowedChannels []*Channel      `json:"channels,omitempty" bun:"m2m:family_allowed_channels,join:Family=Channel"`
	AllowedVideos   []*Video        `json:"allowedVideos,omitempty" bun:"m2m:family_allowed_videos,join:Family=Video"`
	BlockedVideos   []*Video        `json:"blockedVideos,omitempty" bun:"m2m:family_blocked_videos,join:Family=Video"`
}

type FamilyMember struct {
	bun.BaseModel `bun:"table:family_members"`

	FamilyId int32            `json:"familyId" bun:",pk"`
	Family   *Family          `json:"family,omitempty" bun:"rel:belongs-to,join:family_id=id"`
	UserId   int32            `json:"userId" bun:",pk"`
	User     *User            `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Role     FamilyMemberRole `json:"role"`
}

type FamilyAllowedChannel struct {
	bun.BaseModel `bun:"table:family_allowed_channels"`

	FamilyId  int32    `json:"familyId" bun:",pk"`
	Family    *Family  `json:"family,omitempty" bun:"rel:belongs-to,join:family_id=id"`
	ChannelId string   `json:"channelId" bun:",pk"`
	Channel   *Channel `json:"channel,omitempty" bun:"rel:belongs-to,join:channel_id=id"`
}

type FamilyAllowedVideo struct {
	bun.BaseModel `bun:"table:family_allowed_videos"`

	FamilyId int32   `json:"familyId" bun:",pk"`
	Family   *Family `json:"family,omitempty" bun:"rel:belongs-to,join:family_id=id"`
	VideoId  string  `json:"videoId" bun:",pk"`
	Video    *Video  `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}

type FamilyBlockedVideo struct {
	bun.BaseModel `bun:"table:family_blocked_videos"`

	FamilyId int32   `json:"familyId" bun:",pk"`
	Family   *Family `json:"family,omitempty" bun:"rel:belongs-to,join:family_id=id"`
	VideoId  string  `json:"videoId" bun:",pk"`
	Video    *Video  `json:"video,omitempty" bun:"rel:belongs-to,join:video_id=id"`
}
