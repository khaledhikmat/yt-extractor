INSERT INTO videos (
    channel_id, video_id, video_url, title, published_at, duration, short, updated_at,
    views, comments, likes, 
    extraction_url, extracted_at, externalized_at, audio_url, audioed_at, transcription_url, transcribed_at   
) VALUES (
    :channel_id, :video_id, :video_url, :title, :published_at, :duration, :short, NOW(),
    :views, :comments, :likes, 
    :extraction_url, :extracted_at, :externalized_at, :audio_url, :audioed_at, :transcription_url, :transcribed_at
)
RETURNING id
