CREATE OR REPLACE FUNCTION get_channel_title(channel_id TEXT)
RETURNS TEXT
AS $$
DECLARE channel_title TEXT;
BEGIN
    SELECT title INTO channel_title FROM channels WHERE id=get_channel_title.channel_id;
    RETURN channel_title;
END
$$
IMMUTABLE
LANGUAGE plpgsql
RETURNS NULL ON NULL INPUT;

ALTER TABLE videos ADD COLUMN IF NOT EXISTS search_document tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('english', title), 'A') ||
    setweight(to_tsvector('english', description), 'B') ||
    setweight(to_tsvector('english', get_channel_title(channel_id)), 'C')
) STORED;

CREATE INDEX IF NOT EXISTS video_search_index ON videos USING GIN (search_document);