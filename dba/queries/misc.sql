SELECT count(*) from videos where extraction_url is null;
SELECT count(*) from videos;
SELECT count(*) from videos where externalized_at is not null;
SELECT count(*) from videos where externalized_at is null;
SELECT count(*) from videos where processed_at is not null;
SELECT count(*) from videos where audio_url is not null;
SELECT count(*) from videos where published_at >= '2025-01-01 00:00:00';
SELECT count(*) from videos where published_at >= '2025-01-01 00:00:00' AND audio_url is not null;
SELECT * from videos where published_at >= '2025-01-01 00:00:00' AND audio_url is null;
select count(*) from jobs where state = 'completed';
select * from videos order by published_at desc limit 50;
select * from videos where video_url IN ('https://www.youtube.com/watch?v=aUSJG8AI05g', 'https://www.youtube.com/watch?v=J2f8LUZFcD8');
select * from videos where extraction_url = 'https://www.isitdownrightnow.com';
SELECT * from videos where published_at >= '2025-01-01 00:00:00' ORDER BY published_at DESC;
SELECT title, extraction_url, extracted_at, audio_url, audioed_at, transcription_url, transcribed_at from videos WHERE id = 42;
SELECT * from videos where externalized_at is null;
select * from videos where id = 43;
-- update videos SET extraction_url = 'https://www.isitdownrightnow.com' WHERE id = 1219; 
-- update videos SET processed_at = null, externalized_at = null;
select * from jobs where state = 'running';
--delete from jobs where state = 'running';
select * from jobs where id = 590;
select * from jobs ORDER BY STARTED_AT DESC LIMIT 10;
SELECT * FROM jobs WHERE STATE = 'completed' ORDER BY started_at DESC LIMIT 10;
select * from jobs ORDER BY started_at DESC LIMIT 10;
-- delete from jobs where id < 592;
select * from errors;
-- delete from errors;

-- TEST Video ID: 1551
select * from videos ORDER BY ID DESC LIMIT 20;
select * from videos ORDER BY published_at DESC LIMIT 30;
select * from videos where id = 42;

UPDATE videos 
SET 
audioed_at = null, 
audio_url = null
WHERE ID = 42;

UPDATE videos 
SET 
audioed_at = null, 
audio_url = null, 
published_at = '2025-02-01 00:00:00' 
WHERE ID = xxx;

UPDATE videos 
SET 
published_at = '2013-02-09 01:44:03' 
WHERE ID = xxx;

-- Audio Criteria
SELECT * FROM videos 
WHERE channel_id = 'UCP-PfkMcOKriSxFMH7pTxfA' 
AND externalized_at is not null 
AND extracted_at is not null 
AND extraction_url != 'https://www.isitdownrightnow.com' 
AND audioed_at is null 
AND published_at >= '2025-01-01 00:00:00'
ORDER BY published_at DESC 
LIMIT 10

-- Transcription criteria
SELECT * FROM videos 
WHERE channel_id = 'UCP-PfkMcOKriSxFMH7pTxfA' 
AND externalized_at is not null 
AND extracted_at is not null 
AND extraction_url != 'https://www.isitdownrightnow.com'  
AND audioed_at is not null 
AND audio_url != 'https://www.isitdownrightnow.com' 
AND audio_url != 'https://httpstatuses.com/202' 
AND transcribed_at is null 
AND published_at >= '2025-01-01 00:00:00'
ORDER BY published_at DESC 
LIMIT 10

INSERT INTO videos (
    channel_id, video_id, video_url, title, published_at, duration, short, updated_at, externalized_at, 
    views, comments, likes, extraction_url, extracted_at   
) VALUES (
    'UCP-PfkMcOKriSxFMH7pTxfA', 'teststuff', 'https://www.youtube.com/watch?v=8DaUwm7d54o', 'test-try-audio', NOW(), 38, true, NOW(), NOW(),
    10, 10, 10, 'https://yt-extractor.s3.us-east-2.amazonaws.com/UCP-PfkMcOKriSxFMH7pTxfA/8DaUwm7d54o.mp4', NOW()
)



