SELECT
    "video".*,
    "channel".*,
    "video_view"."progress"
FROM
    "users" AS "user" LEFT JOIN
    "allowed_channels" AS "allowed_channel" ON "user"."id" = "allowed_channel"."user_id" LEFT JOIN
    "allowed_videos" AS "allowed_video" ON "user"."id" = "allowed_video"."user_id" LEFT JOIN
    "videos" AS "video" ON "allowed_channel"."channel_id" = "video"."channel_id" OR "allowed_video"."video_id" = "video"."id" LEFT JOIN
    "blocked_videos" AS "blocked_video" ON "user"."id" = "blocked_video"."user_id" AND "video"."id" = "blocked_video"."video_id" LEFT JOIN
    "channels" AS "channel" ON "video"."channel_id" = "channel"."id" LEFT JOIN
    "video_views" AS "video_view" ON "user"."id" = "video_view"."user_id" AND "video"."id" = "video_view"."video_id"
WHERE
    "user"."id" = ? AND
    "blocked_video"."video_id" IS NULL;