INSERT INTO videos (
    channel_id, video_id, video_url, title, published_at, duration, short, updated_at,
    views, comments, likes, 
    extraction_url, extracted_at, externalized_at, processed_at
) VALUES (
    :channel_id, :video_id, :video_url, :title, :published_at, :duration, :short, NOW(),
    :views, :comments, :likes, 
    :extraction_url, :extracted_at, :externalized_at, :processed_at
)
RETURNING id
