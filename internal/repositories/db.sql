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

INSERT INTO categories (id, title) VALUES
	(1, 'Film & Animation'),
	(2, 'Autos & Vehicles'),
	(10, 'Music'),
	(15, 'Pets & Animals'),
	(17, 'Sports'),
	(18, 'Short Movies'),
	(19, 'Travel & Events'),
	(20, 'Gaming'),
	(21, 'Videoblogging'),
	(22, 'People & Blogs'),
	(23, 'Comedy'),
	(24, 'Entertainment'),
	(25, 'News & Politics'),
	(26, 'Howto & Style'),
	(27, 'Education'),
	(28, 'Science & Technology'),
	(29, 'Nonprofits & Activism'),
	(30, 'Movies'),
	(31, 'Anime/Animation'),
	(32, 'Action/Adventure'),
	(33, 'Classics'),
	(34, 'Comedy'),
	(35, 'Documentary'),
	(36, 'Drama'),
	(37, 'Family'),
	(38, 'Foreign'),
	(39, 'Horror'),
	(40, 'Sci-Fi/Fantasy'),
	(41, 'Thriller'),
	(42, 'Shorts'),
	(43, 'Shows'),
	(44, 'Trailers')
ON CONFLICT (id) DO NOTHING;
