INSERT INTO videos (
    channel_id, video_id, video_url, title, published_at, duration, short, updated_at, externalized_at, 
    views, comments, likes, extraction_url, extracted_at   
) VALUES (
    'UCP-PfkMcOKriSxFMH7pTxfA', '8DaUwm7d54o', 'https://www.youtube.com/watch?v=8DaUwm7d54o', 'test-try-audio', NOW(), 38, true, NOW(), NOW(),
    10, 10, 10, 'https://yt-extractor.s3.us-east-2.amazonaws.com/UCP-PfkMcOKriSxFMH7pTxfA/8DaUwm7d54o.mp4', NOW()
)
